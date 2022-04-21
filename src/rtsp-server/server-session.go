// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-08
package rtsp_server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/h264"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/rtph264"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"

	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/base"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/headers"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/ringbuffer"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/rtcpreceiver"
)

type SetupInfo struct {
	TrackID int
	Path    string
	Query   string
	Err     error
}

func setupGetTrackIDPathQuery(newURL *base.URL, mode *headers.TransportMode, aTracks []*ServerSessionAnnouncedTrack, sPath, sQuery *string, sURL *base.URL) (setupInfo SetupInfo) {
	pathAndQuery, ok := newURL.RTSPPathAndQuery()
	if !ok {
		setupInfo.Err = liberrors.ErrorInvalidPath()
		return
	}

	if mode == nil || *mode == headers.TransportModePlay {
		i := stringsReverseIndex(pathAndQuery, "/trackID=")

		// URL doesn't contain TrackID - it's track zero
		if i < 0 {
			if !strings.HasSuffix(pathAndQuery, "/") {
				setupInfo.Err = liberrors.ErrorSetupRequestPathMustEndWithSlash()
				return
			}
			pathAndQuery = pathAndQuery[:len(pathAndQuery)-1]

			path, query := base.PathSplitQuery(pathAndQuery)
			setupInfo = SetupInfo{
				TrackID: 0,
				Path:    path,
				Query:   query,
				Err:     nil,
			}

			// Assume is track 0
			return
		}

		tmp, err := strconv.ParseInt(pathAndQuery[i+len("/trackID="):], 10, 64)
		if err != nil || tmp < 0 {
			setupInfo.Err = liberrors.ErrorUnableToParseTrackID(pathAndQuery)
			return
		}

		trackID := int(tmp)
		pathAndQuery = pathAndQuery[:i]

		path, query := base.PathSplitQuery(pathAndQuery)
		if sPath != nil && (path != *sPath || query != *sQuery) {
			setupInfo.Err = liberrors.ErrorCannotSetupTracksWithDifferentPaths()
			return
		}

		setupInfo = SetupInfo{
			TrackID: trackID,
			Path:    path,
			Query:   query,
			Err:     nil,
		}

		return
	}

	for trackID, track := range aTracks {
		u, _ := track.track.url(sURL)
		if u.String() == newURL.String() {
			setupInfo = SetupInfo{
				TrackID: trackID,
				Path:    *sPath,
				Query:   *sQuery,
				Err:     nil,
			}

			return
		}
	}

	setupInfo = SetupInfo{
		Err: liberrors.ErrorInvalidTrackPath(pathAndQuery),
	}

	return
}

func setupGetTransport(th headers.Transport) (Transport, bool) {
	delivery := func() headers.TransportDelivery {
		if th.Delivery != nil {
			return *th.Delivery
		}
		return headers.TransportDeliveryUnicast
	}()

	switch th.Protocol {
	case headers.TransportProtocolUDP:
		if delivery == headers.TransportDeliveryUnicast {
			return TransportUDP, true
		}
		return TransportUDPMulticast, true

	default: // TCP
		if delivery != headers.TransportDeliveryUnicast {
			return 0, false
		}
		return TransportTCP, true
	}
}

// ServerSessionSetupTrack is a setup track of a ServerSession.
type ServerSessionSetupTrack struct {
	tcpChannel  int
	udpRTPPort  int
	udpRTCPPort int
	udpRTPAddr  *net.UDPAddr
	udpRTCPAddr *net.UDPAddr
}

// ServerSessionAnnouncedTrack is an announced track of a ServerSession.
type ServerSessionAnnouncedTrack struct {
	track        Track
	rtcpReceiver *rtcpreceiver.RTCPReceiver
	h264Decoder  *rtph264.Decoder
	h264Encoder  *rtph264.Encoder
}

// ServerSession is a server-side RTSP session.
type ServerSession struct {
	s        *Server
	secretID string // must not be shared, allows taking ownership of the session
	author   *ServerCxn

	ctx                 context.Context
	ctxCancel           func()
	conns               map[*ServerCxn]struct{}
	state               ServerSessionState
	setupTracks         map[int]*ServerSessionSetupTrack
	tcpTracksByChannel  map[int]int
	setupTransport      *Transport
	setupBaseURL        *base.URL     // publish
	setupStream         *ServerStream // read
	setupPath           *string
	setupQuery          *string
	lastRequestTime     time.Time
	tcpConn             *ServerCxn
	announcedTracks     []*ServerSessionAnnouncedTrack // publish
	udpLastFrameTime    *int64                         // publish
	udpCheckStreamTimer *time.Timer
	writerRunning       bool
	writeBuffer         *ringbuffer.RingBuffer

	// writer channels
	writerDone chan struct{}

	// in
	request     chan sessionRequestReq
	connRemove  chan *ServerCxn
	startWriter chan struct{}
}

func newServerSession(
	s *Server,
	secretID string,
	author *ServerCxn,
) *ServerSession {
	ctx, ctxCancel := context.WithCancel(s.ctx)

	ss := &ServerSession{
		s:                   s,
		secretID:            secretID,
		author:              author,
		ctx:                 ctx,
		ctxCancel:           ctxCancel,
		conns:               make(map[*ServerCxn]struct{}),
		lastRequestTime:     time.Now(),
		udpCheckStreamTimer: emptyTimer(),
		request:             make(chan sessionRequestReq),
		connRemove:          make(chan *ServerCxn),
		startWriter:         make(chan struct{}),
	}

	s.wg.Add(1)
	go ss.run()

	return ss
}

// Close closes the ServerSession.
func (ss *ServerSession) Close() error {
	ss.ctxCancel()
	return nil
}

// State returns the state of the session.
func (ss *ServerSession) State() ServerSessionState {
	return ss.state
}

// SetupTracks returns the setup tracks.
func (ss *ServerSession) SetupTracks() map[int]*ServerSessionSetupTrack {
	return ss.setupTracks
}

// SetupTransport returns the transport of the setup tracks.
func (ss *ServerSession) SetupTransport() *Transport {
	return ss.setupTransport
}

// AnnouncedTracks returns the announced tracks.
func (ss *ServerSession) AnnouncedTracks() []*ServerSessionAnnouncedTrack {
	return ss.announcedTracks
}

func (ss *ServerSession) checkState(allowed map[ServerSessionState]struct{}) error {
	if _, ok := allowed[ss.state]; ok {
		return nil
	}

	allowedList := make([]fmt.Stringer, len(allowed))
	i := 0
	for a := range allowed {
		allowedList[i] = a
		i++
	}

	return liberrors.ErrorInvalidState(allowedList, ss.state)
}

func (ss *ServerSession) run() {
	defer ss.s.wg.Done()

	if h, ok := ss.s.Handler.(ServerHandlerOnSessionOpen); ok {
		h.OnSessionOpen(&ServerHandlerOnSessionOpenCtx{
			Session: ss,
			Conn:    ss.author,
		})
	}

	err := ss.runInner()

	ss.ctxCancel()

	switch ss.state {
	case ServerSessionStatePlay:
		ss.setupStream.readerSetInactive(ss)

		if *ss.setupTransport == TransportUDP {
			ss.s.udpRTCPListener.removeClient(ss)
		}

	case ServerSessionStateRecord:
		if *ss.setupTransport == TransportUDP {
			ss.s.udpRTPListener.removeClient(ss)
			ss.s.udpRTCPListener.removeClient(ss)

			for _, at := range ss.announcedTracks {
				at.rtcpReceiver.Close()
				at.rtcpReceiver = nil
			}
		}
	}

	if ss.setupStream != nil {
		ss.setupStream.readerRemove(ss)
	}

	if ss.writerRunning {
		ss.writeBuffer.Close()
		<-ss.writerDone
	}

	for sc := range ss.conns {
		if sc == ss.tcpConn {
			sc.Close()

			// make sure that OnFrame() is never called after OnSessionClose()
			<-sc.done
		}

		select {
		case sc.sessionRemove <- ss:
		case <-sc.ctx.Done():
		}
	}

	select {
	case ss.s.sessionClose <- ss:
	case <-ss.s.ctx.Done():
	}

	if h, ok := ss.s.Handler.(ServerHandlerOnSessionClose); ok {
		h.OnSessionClose(&ServerHandlerOnSessionCloseCtx{
			Session: ss,
			Error:   err,
		})
	}
}

func (ss *ServerSession) runInner() error {
	for {
		select {
		case req := <-ss.request:
			ss.lastRequestTime = time.Now()

			if _, ok := ss.conns[req.sc]; !ok {
				ss.conns[req.sc] = struct{}{}
			}

			res, err := ss.handleRequest(req.sc, req.req)

			var returnedSession *ServerSession
			if err == nil || err == errSwitchReadFunc {
				// ANNOUNCE responses don't contain the session header.
				if req.req.Method != base.Announce &&
					req.req.Method != base.Teardown {
					if res.Header == nil {
						res.Header = make(base.Header)
					}

					res.Header["Session"] = headers.Session{
						Session: ss.secretID,
						Timeout: func() *uint {
							// timeout controls sending of RTCP keepalive.
							// These are needed only when the client is playing
							// and transport is UDP or UDP-multicast.
							if (ss.state == ServerSessionStatePrePlay ||
								ss.state == ServerSessionStatePlay) &&
								(*ss.setupTransport == TransportUDP ||
									*ss.setupTransport == TransportUDPMulticast) {
								v := uint(ss.s.sessionTimeout / time.Second)
								return &v
							}
							return nil
						}(),
					}.Write()
				}

				// after a TEARDOWN, session must be unpaired with the connection.
				if req.req.Method != base.Teardown {
					returnedSession = ss
				}
			}

			savedMethod := req.req.Method

			req.res <- sessionRequestRes{
				res: res,
				err: err,
				ss:  returnedSession,
			}

			if (err == nil || err == errSwitchReadFunc) && savedMethod == base.Teardown {
				return liberrors.ErrorSessionTeardown(req.sc.NetConn().RemoteAddr())
			}

		case sc := <-ss.connRemove:
			delete(ss.conns, sc)

			// if session is not in state RECORD or PLAY, or transport is TCP,
			// and there are no associated connections,
			// close the session.
			if ((ss.state != ServerSessionStateRecord &&
				ss.state != ServerSessionStatePlay) ||
				*ss.setupTransport == TransportTCP) &&
				len(ss.conns) == 0 {
				return liberrors.ErrorSessionNotInUse()
			}

		case <-ss.startWriter:
			if !ss.writerRunning && (ss.state == ServerSessionStateRecord ||
				ss.state == ServerSessionStatePlay) &&
				*ss.setupTransport == TransportTCP {
				ss.writerRunning = true
				ss.writerDone = make(chan struct{})
				go ss.runWriter()
			}

		case <-ss.udpCheckStreamTimer.C:
			now := time.Now()

			// in case of RECORD, timeout happens when no RTP or RTCP packets are being received
			if ss.state == ServerSessionStateRecord {
				lft := atomic.LoadInt64(ss.udpLastFrameTime)
				if now.Sub(time.Unix(lft, 0)) >= ss.s.ReadTimeout {
					return liberrors.ErrorNoUDPPacketsInAWhile()
				}

				// in case of PLAY, timeout happens when no RTSP keepalive are being received
			} else if now.Sub(ss.lastRequestTime) >= ss.s.sessionTimeout {
				return liberrors.ErrorNoRTSPRequestsInAWhile()
			}

			ss.udpCheckStreamTimer = time.NewTimer(ss.s.checkStreamPeriod)

		case <-ss.ctx.Done():
			return liberrors.ErrorServerTerminated()
		}
	}
}

func (ss *ServerSession) handleRequest(sc *ServerCxn, req *base.Request) (*base.Response, error) {
	if ss.tcpConn != nil && sc != ss.tcpConn {
		return &base.Response{
			StatusCode: base.StatusBadRequest,
		}, liberrors.ErrorSessionLinkedToOtherConn()
	}

	switch req.Method {
	case base.Options:
		var methods []string
		if _, ok := sc.s.Handler.(ServerHandlerOnDescribe); ok {
			methods = append(methods, string(base.Describe))
		}
		if _, ok := sc.s.Handler.(ServerHandlerOnAnnounce); ok {
			methods = append(methods, string(base.Announce))
		}
		if _, ok := sc.s.Handler.(ServerHandlerOnSetup); ok {
			methods = append(methods, string(base.Setup))
		}
		if _, ok := sc.s.Handler.(ServerHandlerOnPlay); ok {
			methods = append(methods, string(base.Play))
		}
		if _, ok := sc.s.Handler.(ServerHandlerOnRecord); ok {
			methods = append(methods, string(base.Record))
		}
		if _, ok := sc.s.Handler.(ServerHandlerOnPause); ok {
			methods = append(methods, string(base.Pause))
		}
		methods = append(methods, string(base.GetParameter))
		if _, ok := sc.s.Handler.(ServerHandlerOnSetParameter); ok {
			methods = append(methods, string(base.SetParameter))
		}
		methods = append(methods, string(base.Teardown))

		return &base.Response{
			StatusCode: base.StatusOK,
			Header: base.Header{
				"Public": base.HeaderValue{strings.Join(methods, ", ")},
			},
		}, nil

	case base.Announce:
		err := ss.checkState(map[ServerSessionState]struct{}{
			ServerSessionStateInitial: {},
		})
		if err != nil {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, err
		}

		pathAndQuery, ok := req.URL.RTSPPathAndQuery()
		if !ok {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorInvalidPath()
		}

		path, query := base.PathSplitQuery(pathAndQuery)

		ct, ok := req.Header["Content-Type"]
		if !ok || len(ct) != 1 {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorContentTypeMissing()
		}

		if ct[0] != "application/sdp" {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorContentTypeUnsupported(ct[0])
		}

		tracks, err := ReadTracks(req.Body, false)
		if err != nil {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorSDPInvalid(err)
		}

		for _, track := range tracks {
			trackURL, err := track.url(req.URL)
			if err != nil {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorUnableToGenerateTrackURL()
			}

			trackPath, ok := trackURL.RTSPPathAndQuery()
			if !ok {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorInvalidTrackURL(trackURL.String())
			}

			if !strings.HasPrefix(trackPath, path) {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorInvalidTrackPathMustBeginWith(path, trackPath)
			}
		}

		res, err := ss.s.Handler.(ServerHandlerOnAnnounce).OnAnnounce(&ServerHandlerOnAnnounceCtx{
			Server:  ss.s,
			Session: ss,
			Conn:    sc,
			Request: req,
			Path:    path,
			Query:   query,
			Tracks:  tracks,
		})

		if res.StatusCode != base.StatusOK {
			return res, err
		}

		ss.state = ServerSessionStatePreRecord
		ss.setupPath = &path
		ss.setupQuery = &query
		ss.setupBaseURL = req.URL

		ss.announcedTracks = make([]*ServerSessionAnnouncedTrack, len(tracks))
		for trackID, track := range tracks {
			ss.announcedTracks[trackID] = &ServerSessionAnnouncedTrack{
				track: track,
			}
		}

		v := time.Now().Unix()
		ss.udpLastFrameTime = &v
		return res, err

	case base.Setup:
		err := ss.checkState(map[ServerSessionState]struct{}{
			ServerSessionStateInitial:   {},
			ServerSessionStatePrePlay:   {},
			ServerSessionStatePreRecord: {},
		})
		if err != nil {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, err
		}

		var inTH headers.Transport
		err = inTH.Read(req.Header["Transport"])
		if err != nil {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorTransportHeaderInvalid(err)
		}

		setupInfo := setupGetTrackIDPathQuery(
			req.URL,
			inTH.Mode,
			ss.announcedTracks,
			ss.setupPath,
			ss.setupQuery,
			ss.setupBaseURL,
		)
		if err != nil {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, err
		}
		if _, ok := ss.setupTracks[setupInfo.TrackID]; ok {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorTrackAlreadySetup(setupInfo.TrackID)
		}

		transport, ok := setupGetTransport(inTH)
		if !ok {
			return &base.Response{
				StatusCode: base.StatusUnsupportedTransport,
			}, nil
		}

		switch transport {
		case TransportUDP:
			if inTH.ClientPorts == nil {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorTransportHeaderNoClientPorts()
			}

			if ss.s.udpRTPListener == nil {
				return &base.Response{
					StatusCode: base.StatusUnsupportedTransport,
				}, nil
			}

		case TransportUDPMulticast:
			if ss.s.MulticastIPRange == "" {
				return &base.Response{
					StatusCode: base.StatusUnsupportedTransport,
				}, nil
			}

		default: // TCP
			if inTH.InterleavedIDs == nil {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorTransportHeaderNoInterleavedIDs()
			}

			if (inTH.InterleavedIDs[0]%2) != 0 ||
				(inTH.InterleavedIDs[0]+1) != inTH.InterleavedIDs[1] {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorTransportHeaderInvalidInterleavedIDs()
			}

			if _, ok := ss.tcpTracksByChannel[inTH.InterleavedIDs[0]]; ok {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorTransportHeaderInterleavedIDsAlreadyUsed()
			}
		}

		if ss.setupTransport != nil && *ss.setupTransport != transport {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorTracksDifferentProtocols()
		}

		switch ss.state {
		case ServerSessionStateInitial, ServerSessionStatePrePlay: // play
			if inTH.Mode != nil && *inTH.Mode != headers.TransportModePlay {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorTransportHeaderInvalidMode(inTH.Mode.AsInt())
			}

		default: // record
			if transport == TransportUDPMulticast {
				return &base.Response{
					StatusCode: base.StatusUnsupportedTransport,
				}, nil
			}

			if inTH.Mode == nil || *inTH.Mode != headers.TransportModeRecord {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorTransportHeaderInvalidMode(inTH.Mode.AsInt())
			}
		}

		res, stream, err := ss.s.Handler.(ServerHandlerOnSetup).OnSetup(&ServerHandlerOnSetupCtx{
			Server:    ss.s,
			Session:   ss,
			Conn:      sc,
			Request:   req,
			Path:      setupInfo.Path,
			Query:     setupInfo.Query,
			TrackID:   setupInfo.TrackID,
			Transport: transport,
		})

		// workaround to prevent a bug in rtsp-client-sink
		// that makes impossible for the client to receive the response
		// and send frames.
		// this was causing problems during unit tests.
		if ua, ok := req.Header["User-Agent"]; ok && len(ua) == 1 &&
			strings.HasPrefix(ua[0], "GStreamer") {
			select {
			case <-time.After(1 * time.Second):
			case <-ss.ctx.Done():
			}
		}

		if res.StatusCode != base.StatusOK {
			return res, err
		}

		if ss.state == ServerSessionStateInitial {
			err := stream.readerAdd(ss,
				transport,
				inTH.ClientPorts,
			)
			if err != nil {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, err
			}

			ss.state = ServerSessionStatePrePlay
			ss.setupPath = &setupInfo.Path
			ss.setupQuery = &setupInfo.Query
			ss.setupStream = stream
		}

		th := headers.Transport{}

		if ss.state == ServerSessionStatePrePlay {
			ssrc := stream.ssrc(setupInfo.TrackID)
			if ssrc != 0 {
				th.SSRC = &ssrc
			}
		}

		ss.setupTransport = &transport

		if res.Header == nil {
			res.Header = make(base.Header)
		}

		sst := &ServerSessionSetupTrack{}

		switch transport {
		case TransportUDP:
			sst.udpRTPPort = inTH.ClientPorts[0]
			sst.udpRTCPPort = inTH.ClientPorts[1]

			sst.udpRTPAddr = &net.UDPAddr{
				IP:   ss.author.ip(),
				Zone: ss.author.zone(),
				Port: sst.udpRTPPort,
			}

			sst.udpRTCPAddr = &net.UDPAddr{
				IP:   ss.author.ip(),
				Zone: ss.author.zone(),
				Port: sst.udpRTCPPort,
			}

			th.Protocol = headers.TransportProtocolUDP
			de := headers.TransportDeliveryUnicast
			th.Delivery = &de
			th.ClientPorts = inTH.ClientPorts
			th.ServerPorts = &[2]int{sc.s.udpRTPListener.port(), sc.s.udpRTCPListener.port()}

		case TransportUDPMulticast:
			th.Protocol = headers.TransportProtocolUDP
			de := headers.TransportDeliveryMulticast
			th.Delivery = &de
			v := uint(127)
			th.TTL = &v
			d := stream.serverMulticastHandlers[setupInfo.TrackID].ip()
			th.Destination = &d
			th.Ports = &[2]int{ss.s.MulticastRTPPort, ss.s.MulticastRTCPPort}

		default: // TCP
			sst.tcpChannel = inTH.InterleavedIDs[0]

			if ss.tcpTracksByChannel == nil {
				ss.tcpTracksByChannel = make(map[int]int)
			}

			ss.tcpTracksByChannel[inTH.InterleavedIDs[0]] = setupInfo.TrackID

			th.Protocol = headers.TransportProtocolTCP
			de := headers.TransportDeliveryUnicast
			th.Delivery = &de
			th.InterleavedIDs = inTH.InterleavedIDs
		}

		if ss.setupTracks == nil {
			ss.setupTracks = make(map[int]*ServerSessionSetupTrack)
		}

		ss.setupTracks[setupInfo.TrackID] = sst

		res.Header["Transport"] = th.Write()

		return res, err

	case base.Play:
		// play can be sent twice, allow calling it even if we're already playing
		err := ss.checkState(map[ServerSessionState]struct{}{
			ServerSessionStatePrePlay: {},
			ServerSessionStatePlay:    {},
		})
		if err != nil {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, err
		}

		pathAndQuery, ok := req.URL.RTSPPathAndQuery()
		if !ok {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorInvalidPath()
		}

		// Path can end with a slash due to Content-Base, remove it
		pathAndQuery = strings.TrimSuffix(pathAndQuery, "/")

		path, query := base.PathSplitQuery(pathAndQuery)

		if ss.State() == ServerSessionStatePrePlay &&
			path != *ss.setupPath {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorServerPathHasChanged(*ss.setupPath, path)
		}

		// Allocate writeBuffer before calling OnPlay().
		// in this way it's possible to call ServerSession.WritePacket*()
		// inside the callback.
		if ss.state != ServerSessionStatePlay &&
			*ss.setupTransport != TransportUDPMulticast {
			ss.writeBuffer = ringbuffer.New(uint64(ss.s.WriteBufferCount))
		}

		res, err := sc.s.Handler.(ServerHandlerOnPlay).OnPlay(&ServerHandlerOnPlayCtx{
			Session: ss,
			Conn:    sc,
			Request: req,
			Path:    path,
			Query:   query,
		})

		if res.StatusCode != base.StatusOK {
			if ss.state != ServerSessionStatePlay {
				ss.writeBuffer = nil
			}
			return res, err
		}

		if ss.state == ServerSessionStatePlay {
			return res, err
		}

		ss.state = ServerSessionStatePlay

		switch *ss.setupTransport {
		case TransportUDP:
			ss.udpCheckStreamTimer = time.NewTimer(ss.s.checkStreamPeriod)

			ss.writerRunning = true
			ss.writerDone = make(chan struct{})
			go ss.runWriter()

			for trackID, track := range ss.setupTracks {
				// readers can send RTCP packets only
				sc.s.udpRTCPListener.addClient(ss.author.ip(), track.udpRTCPPort, ss, trackID, false)

				// firewall opening is performed by RTCP sender reports generated by ServerStream
			}

		case TransportUDPMulticast:
			ss.udpCheckStreamTimer = time.NewTimer(ss.s.checkStreamPeriod)

		default: // TCP
			ss.tcpConn = sc
			ss.tcpConn.readFunc = ss.tcpConn.readFuncTCP
			err = errSwitchReadFunc

			// runWriter() is called by ServerCxn after the response has been sent
		}

		ss.setupStream.readerSetActive(ss)

		// add RTP-Info
		var trackIDs []int
		for trackID := range ss.setupTracks {
			trackIDs = append(trackIDs, trackID)
		}
		sort.Slice(trackIDs, func(a, b int) bool {
			return trackIDs[a] < trackIDs[b]
		})
		var ri headers.RTPInfo
		for _, trackID := range trackIDs {
			ts := ss.setupStream.timestamp(trackID)
			if ts == 0 {
				continue
			}

			u := &base.URL{
				Scheme: req.URL.Scheme,
				User:   req.URL.User,
				Host:   req.URL.Host,
				Path:   "/" + *ss.setupPath + "/trackID=" + strconv.FormatInt(int64(trackID), 10),
			}

			lsn := ss.setupStream.lastSequenceNumber(trackID)

			ri = append(ri, &headers.RTPInfoEntry{
				URL:            u.String(),
				SequenceNumber: &lsn,
				Timestamp:      &ts,
			})
		}
		if len(ri) > 0 {
			if res.Header == nil {
				res.Header = make(base.Header)
			}
			res.Header["RTP-Info"] = ri.Write()
		}

		return res, err

	case base.Record:
		err := ss.checkState(map[ServerSessionState]struct{}{
			ServerSessionStatePreRecord: {},
		})
		if err != nil {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, err
		}

		if len(ss.setupTracks) != len(ss.announcedTracks) {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorNotAllAnnouncedTracksSetup()
		}

		pathAndQuery, ok := req.URL.RTSPPathAndQuery()
		if !ok {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorInvalidPath()
		}

		// Path can end with a slash due to Content-Base, remove it
		pathAndQuery = strings.TrimSuffix(pathAndQuery, "/")

		path, query := base.PathSplitQuery(pathAndQuery)

		if path != *ss.setupPath {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorServerPathHasChanged(*ss.setupPath, path)
		}

		// allocate writeBuffer before calling OnRecord().
		// in this way it's possible to call ServerSession.WritePacket*()
		// inside the callback.
		// when recording, writeBuffer is only used to send RTCP receiver reports,
		// that are much smaller than RTP packets and are sent at a fixed interval.
		// decrease RAM consumption by allocating fewer buffers.
		ss.writeBuffer = ringbuffer.New(uint64(8))

		res, err := ss.s.Handler.(ServerHandlerOnRecord).OnRecord(&ServerHandlerOnRecordCtx{
			Session: ss,
			Conn:    sc,
			Request: req,
			Path:    path,
			Query:   query,
		})

		if res.StatusCode != base.StatusOK {
			ss.writeBuffer = nil
			return res, err
		}

		ss.state = ServerSessionStateRecord

		for _, at := range ss.announcedTracks {
			if _, ok := at.track.(*TrackH264); ok {
				at.h264Decoder = &rtph264.Decoder{}
				at.h264Decoder.Init()
			}
		}

		switch *ss.setupTransport {
		case TransportUDP:
			ss.udpCheckStreamTimer = time.NewTimer(ss.s.checkStreamPeriod)

			ss.writerRunning = true
			ss.writerDone = make(chan struct{})
			go ss.runWriter()

			for trackID, at := range ss.announcedTracks {
				// open the firewall by sending packets to the counterpart
				ss.WritePacketRTP(trackID, &rtp.Packet{Header: rtp.Header{Version: 2}})
				ss.WritePacketRTCP(trackID, &rtcp.ReceiverReport{})

				cTrackID := trackID

				at.rtcpReceiver = rtcpreceiver.New(ss.s.udpReceiverReportPeriod,
					nil, at.track.ClockRate(), func(pkt rtcp.Packet) {
						ss.WritePacketRTCP(cTrackID, pkt)
					})

				ss.s.udpRTPListener.addClient(ss.author.ip(), ss.setupTracks[trackID].udpRTPPort, ss, trackID, true)
				ss.s.udpRTCPListener.addClient(ss.author.ip(), ss.setupTracks[trackID].udpRTCPPort, ss, trackID, true)
			}

		default: // TCP
			ss.tcpConn = sc
			ss.tcpConn.readFunc = ss.tcpConn.readFuncTCP
			err = errSwitchReadFunc

			// runWriter() is called by cxn after sending the response
		}

		return res, err

	case base.Pause:
		err := ss.checkState(map[ServerSessionState]struct{}{
			ServerSessionStatePrePlay:   {},
			ServerSessionStatePlay:      {},
			ServerSessionStatePreRecord: {},
			ServerSessionStateRecord:    {},
		})
		if err != nil {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, err
		}

		pathAndQuery, ok := req.URL.RTSPPathAndQuery()
		if !ok {
			return &base.Response{
				StatusCode: base.StatusBadRequest,
			}, liberrors.ErrorInvalidPath()
		}

		// Path can end with a slash due to Content-Base, remove it
		pathAndQuery = strings.TrimSuffix(pathAndQuery, "/")

		path, query := base.PathSplitQuery(pathAndQuery)

		res, err := ss.s.Handler.(ServerHandlerOnPause).OnPause(&ServerHandlerOnPauseCtx{
			Session: ss,
			Conn:    sc,
			Request: req,
			Path:    path,
			Query:   query,
		})

		if res.StatusCode != base.StatusOK {
			return res, err
		}

		if ss.writerRunning {
			ss.writeBuffer.Close()
			<-ss.writerDone
			ss.writerRunning = false
		}

		switch ss.state {
		case ServerSessionStatePlay:
			ss.setupStream.readerSetInactive(ss)

			ss.state = ServerSessionStatePrePlay

			switch *ss.setupTransport {
			case TransportUDP:
				ss.udpCheckStreamTimer = emptyTimer()

				ss.s.udpRTCPListener.removeClient(ss)

			case TransportUDPMulticast:
				ss.udpCheckStreamTimer = emptyTimer()

			default: // TCP
				ss.tcpConn.readFunc = ss.tcpConn.readFuncStandard
				err = errSwitchReadFunc

				ss.tcpConn = nil
			}

		case ServerSessionStateRecord:

			switch *ss.setupTransport {
			case TransportUDP:
				ss.udpCheckStreamTimer = emptyTimer()

				ss.s.udpRTPListener.removeClient(ss)
				ss.s.udpRTCPListener.removeClient(ss)

				for _, at := range ss.announcedTracks {
					at.rtcpReceiver.Close()
					at.rtcpReceiver = nil
				}

			default: // TCP
				ss.tcpConn.readFunc = ss.tcpConn.readFuncStandard
				err = errSwitchReadFunc
				ss.tcpConn = nil
			}

			for _, at := range ss.announcedTracks {
				at.h264Decoder = nil
				at.h264Encoder = nil
			}

			ss.state = ServerSessionStatePreRecord
		}

		return res, err

	case base.Teardown:
		var err error
		if (ss.state == ServerSessionStatePlay || ss.state == ServerSessionStateRecord) &&
			*ss.setupTransport == TransportTCP {
			ss.tcpConn.readFunc = ss.tcpConn.readFuncStandard
			err = errSwitchReadFunc
		}

		return &base.Response{
			StatusCode: base.StatusOK,
		}, err

	case base.GetParameter:
		if h, ok := sc.s.Handler.(ServerHandlerOnGetParameter); ok {
			pathAndQuery, ok := req.URL.RTSPPathAndQuery()
			if !ok {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorInvalidPath()
			}

			path, query := base.PathSplitQuery(pathAndQuery)

			return h.OnGetParameter(&ServerHandlerOnGetParameterCtx{
				Session: ss,
				Conn:    sc,
				Request: req,
				Path:    path,
				Query:   query,
			})
		}

		// GET_PARAMETER is used like a ping when reading, and sometimes
		// also when publishing; reply with 200
		return &base.Response{
			StatusCode: base.StatusOK,
			Header: base.Header{
				"Content-Type": base.HeaderValue{"text/parameters"},
			},
			Body: []byte{},
		}, nil
	}

	return &base.Response{
		StatusCode: base.StatusBadRequest,
	}, liberrors.ErrorUnhandledRequest(req.Method.String(), req.URL.String())
}

func (ss *ServerSession) runWriter() {
	defer close(ss.writerDone)

	var writeFunc func(int, bool, []byte)

	if *ss.setupTransport == TransportUDP {
		writeFunc = func(trackID int, isRTP bool, payload []byte) {
			if isRTP {
				ss.s.udpRTPListener.write(payload, ss.setupTracks[trackID].udpRTPAddr)
			} else {
				ss.s.udpRTCPListener.write(payload, ss.setupTracks[trackID].udpRTCPAddr)
			}
		}
	} else { // TCP
		rtpFrames := make(map[int]*base.InterleavedFrame, len(ss.setupTracks))
		rtcpFrames := make(map[int]*base.InterleavedFrame, len(ss.setupTracks))

		for trackID, sst := range ss.setupTracks {
			rtpFrames[trackID] = &base.InterleavedFrame{Channel: sst.tcpChannel}
			rtcpFrames[trackID] = &base.InterleavedFrame{Channel: sst.tcpChannel + 1}
		}

		var buf bytes.Buffer

		writeFunc = func(trackID int, isRTP bool, payload []byte) {
			if isRTP {
				f := rtpFrames[trackID]
				f.Payload = payload
				f.Write(&buf)

				ss.tcpConn.cxn.SetWriteDeadline(time.Now().Add(ss.s.WriteTimeout))
				ss.tcpConn.cxn.Write(buf.Bytes())
			} else {
				f := rtcpFrames[trackID]
				f.Payload = payload
				f.Write(&buf)

				ss.tcpConn.cxn.SetWriteDeadline(time.Now().Add(ss.s.WriteTimeout))
				ss.tcpConn.cxn.Write(buf.Bytes())
			}
		}
	}

	for {
		tmp, ok := ss.writeBuffer.Pull()
		if !ok {
			return
		}
		data := tmp.(trackTypePayload)

		writeFunc(data.trackID, data.isRTP, data.payload)
	}
}

func (ss *ServerSession) processPacketRTP(at *ServerSessionAnnouncedTrack, ctx *ServerHandlerOnPacketRTPCtx) {
	// remove padding
	ctx.Packet.Header.Padding = false
	ctx.Packet.PaddingSize = 0

	// decode
	if at.h264Decoder != nil {
		nalus, pts, err := at.h264Decoder.DecodeUntilMarker(ctx.Packet)
		if err == nil {
			ctx.PTSEqualsDTS = h264.IDRPresent(nalus)
			ctx.H264NALUs = nalus
			ctx.H264PTS = pts
		} else {
			ctx.PTSEqualsDTS = false
		}
	} else {
		ctx.PTSEqualsDTS = false
	}
}

func (ss *ServerSession) onPacketRTCP(trackID int, pkt rtcp.Packet) {
	if h, ok := ss.s.Handler.(ServerHandlerOnPacketRTCP); ok {
		h.OnPacketRTCP(&ServerHandlerOnPacketRTCPCtx{
			Session: ss,
			TrackID: trackID,
			Packet:  pkt,
		})
	}
}

func (ss *ServerSession) writePacketRTP(trackID int, bts []byte) {
	if _, ok := ss.setupTracks[trackID]; !ok {
		return
	}

	ss.writeBuffer.Push(trackTypePayload{
		trackID: trackID,
		isRTP:   true,
		payload: bts,
	})
}

// WritePacketRTP writes a RTP packet to the session.
func (ss *ServerSession) WritePacketRTP(trackID int, pkt *rtp.Packet) {
	bts, err := pkt.Marshal()
	if err != nil {
		return
	}

	ss.writePacketRTP(trackID, bts)
}

func (ss *ServerSession) writePacketRTCP(trackID int, bts []byte) {
	if _, ok := ss.setupTracks[trackID]; !ok {
		return
	}

	ss.writeBuffer.Push(trackTypePayload{
		trackID: trackID,
		isRTP:   false,
		payload: bts,
	})
}

// WritePacketRTCP writes a RTCP packet to the session.
func (ss *ServerSession) WritePacketRTCP(trackID int, pkt rtcp.Packet) {
	bts, err := pkt.Marshal()
	if err != nil {
		return
	}

	ss.writePacketRTCP(trackID, bts)
}
