// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-08
package rtsp_server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/rtph264"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/pion/rtcp"

	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/base"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
)

func getSessionID(header base.Header) string {
	if h, ok := header[base.HeaderSession]; ok && len(h) == 1 {
		return h[0]
	}
	return ""
}

type reqWrapper struct {
	req *base.Request
	res chan error
}

// ServerCxn is a server-side RTSP connection.
type ServerCxn struct {
	s   *Server
	cxn net.Conn

	ctx        context.Context
	ctxCancel  func()
	remoteAddr *net.TCPAddr
	reader     *bufio.Reader
	session    *ServerSession
	readFunc   func(reqWrappers chan reqWrapper) error

	// in
	sessionRemove chan *ServerSession

	// out
	done chan struct{}
}

func newServerCxn(s *Server, c net.Conn) *ServerCxn {
	ctx, ctxCancel := context.WithCancel(s.ctx)

	cxn := func() net.Conn {
		if s.TLSConfig != nil {
			return tls.Server(c, s.TLSConfig)
		}
		return c
	}()

	sc := &ServerCxn{
		s:             s,
		cxn:           cxn,
		ctx:           ctx,
		ctxCancel:     ctxCancel,
		remoteAddr:    cxn.RemoteAddr().(*net.TCPAddr),
		sessionRemove: make(chan *ServerSession),
		done:          make(chan struct{}),
	}

	sc.readFunc = sc.readFuncStandard

	s.wg.Add(1)
	go sc.run()

	return sc
}

// Close closes the ServerCxn.
func (sc *ServerCxn) Close() error {
	sc.ctxCancel()
	return nil
}

// NetConn returns the underlying net.Conn.
func (sc *ServerCxn) NetConn() net.Conn {
	return sc.cxn
}

func (sc *ServerCxn) ip() net.IP {
	return sc.remoteAddr.IP
}

func (sc *ServerCxn) zone() string {
	return sc.remoteAddr.Zone
}

func (sc *ServerCxn) run() {
	defer sc.s.wg.Done()
	defer close(sc.done)

	if h, ok := sc.s.Handler.(ServerHandlerOnConnOpen); ok {
		h.OnConnOpen(&ServerHandlerOnConnOpenCtx{
			Conn: sc,
		})
	}

	sc.reader = bufio.NewReaderSize(sc.cxn, tcpReadBufferSize)

	reqWrappers := make(chan reqWrapper)
	readErr := make(chan error)
	readDone := make(chan struct{})
	go sc.runReader(reqWrappers, readErr, readDone)

	err := sc.runInner(reqWrappers, readErr)

	sc.ctxCancel()

	sc.cxn.Close()
	<-readDone

	if sc.session != nil {
		select {
		case sc.session.connRemove <- sc:
		case <-sc.session.ctx.Done():
		}
	}

	select {
	case sc.s.connClose <- sc:
	case <-sc.s.ctx.Done():
	}

	if h, ok := sc.s.Handler.(ServerHandlerOnConnClose); ok {
		h.OnConnClose(&ServerHandlerOnConnCloseCtx{
			Conn:  sc,
			Error: err,
		})
	}
}

func (sc *ServerCxn) runInner(reqWrappers chan reqWrapper, readErr chan error) error {
	for {
		select {
		case wrapper := <-reqWrappers:
			wrapper.res <- sc.handleRequestOuter(wrapper.req)

		case err := <-readErr:
			return err

		case ss := <-sc.sessionRemove:
			if sc.session == ss {
				sc.session = nil
			}

		case <-sc.ctx.Done():
			return liberrors.ErrorServerTerminated()
		}
	}
}

var errSwitchReadFunc = errors.New("switch read function")

func (sc *ServerCxn) runReader(reqWrappers chan reqWrapper, readErr chan error, readDone chan struct{}) {
	defer close(readDone)

	for {
		err := sc.readFunc(reqWrappers)
		if err == errSwitchReadFunc {
			continue
		}

		select {
		case readErr <- err:
		case <-sc.ctx.Done():
		}
		break
	}
}

func (sc *ServerCxn) readFuncStandard(reqWrappers chan reqWrapper) error {
	var req base.Request

	// reset deadline
	_ = sc.cxn.SetReadDeadline(time.Time{})

	for {
		err := req.Read(sc.reader)
		if err != nil {
			return err
		}

		reqErr := make(chan error)
		select {
		case reqWrappers <- reqWrapper{req: &req, res: reqErr}:
			err = <-reqErr
			if err != nil {
				return err
			}

		case <-sc.ctx.Done():
			return liberrors.ErrorServerTerminated()
		}
	}
}

func (sc *ServerCxn) readFuncTCP(readRequest chan reqWrapper) error {
	// reset deadline
	_ = sc.cxn.SetReadDeadline(time.Time{})

	select {
	case sc.session.startWriter <- struct{}{}:
	case <-sc.session.ctx.Done():
	}

	var processFunc func(int, bool, []byte) error

	if sc.session.state == ServerSessionStatePlay {
		processFunc = func(trackID int, isRTP bool, payload []byte) error {
			if !isRTP {
				if len(payload) > maxPacketSize {
					return ErrorPayloadSizeGreaterThanMaximum(len(payload), maxPacketSize)
				}

				packets, err := rtcp.Unmarshal(payload)
				if err != nil {
					return err
				}

				if h, ok := sc.s.Handler.(ServerHandlerOnPacketRTCP); ok {
					for _, packet := range packets {
						h.OnPacketRTCP(&ServerHandlerOnPacketRTCPCtx{
							Session: sc.session,
							TrackID: trackID,
							Packet:  packet,
						})
					}
				}
			}

			return nil
		}
	} else {
		tcpRTPPacketBuffer := newRTPPacketMultiBuffer(uint64(sc.s.ReadBufferCount))

		processFunc = func(trackID int, isRTP bool, payload []byte) error {
			if isRTP {
				pkt := tcpRTPPacketBuffer.next()
				err := pkt.Unmarshal(payload)
				if err != nil {
					return err
				}

				ctx := ServerHandlerOnPacketRTPCtx{
					Session: sc.session,
					TrackID: trackID,
					Packet:  pkt,
				}
				at := sc.session.announcedTracks[trackID]
				sc.session.processPacketRTP(at, &ctx)

				if at.h264Decoder != nil {
					if at.h264Encoder == nil && len(payload) > maxPacketSize {
						v1 := pkt.SSRC
						v2 := pkt.SequenceNumber
						v3 := pkt.Timestamp
						at.h264Encoder = &rtph264.Encoder{
							PayloadType:           pkt.PayloadType,
							SSRC:                  &v1,
							InitialSequenceNumber: &v2,
							InitialTimestamp:      &v3,
						}
						at.h264Encoder.Init()
					}

					if at.h264Encoder != nil {
						if ctx.H264NALUs != nil {
							packets, err := at.h264Encoder.Encode(ctx.H264NALUs, ctx.H264PTS)
							if err != nil {
								return err
							}

							for i, pkt := range packets {
								if i != len(packets)-1 {
									if h, ok := sc.s.Handler.(ServerHandlerOnPacketRTP); ok {
										h.OnPacketRTP(&ServerHandlerOnPacketRTPCtx{
											Session:      sc.session,
											TrackID:      trackID,
											Packet:       pkt,
											PTSEqualsDTS: false,
										})
									}
								} else {
									ctx.Packet = pkt
									if h, ok := sc.s.Handler.(ServerHandlerOnPacketRTP); ok {
										h.OnPacketRTP(&ctx)
									}
								}
							}
						}
					} else {
						if h, ok := sc.s.Handler.(ServerHandlerOnPacketRTP); ok {
							h.OnPacketRTP(&ctx)
						}
					}
				} else {
					if len(payload) > maxPacketSize {
						return ErrorPayloadSizeGreaterThanMaximum(len(payload), maxPacketSize)
					}

					if h, ok := sc.s.Handler.(ServerHandlerOnPacketRTP); ok {
						h.OnPacketRTP(&ctx)
					}
				}
			} else {
				if len(payload) > maxPacketSize {
					return ErrorPayloadSizeGreaterThanMaximum(len(payload), maxPacketSize)
				}

				packets, err := rtcp.Unmarshal(payload)
				if err != nil {
					return err
				}

				for _, pkt := range packets {
					sc.session.onPacketRTCP(trackID, pkt)
				}
			}

			return nil
		}
	}

	var req base.Request
	var frame base.InterleavedFrame

	for {
		if sc.session.state == ServerSessionStateRecord {
			_ = sc.cxn.SetReadDeadline(time.Now().Add(sc.s.ReadTimeout))
		}

		what, err := base.ReadInterleavedFrameOrRequest(&frame, tcpMaxFramePayloadSize, &req, sc.reader)
		if err != nil {
			return err
		}

		switch what.(type) {
		case *base.InterleavedFrame:
			channel := frame.Channel
			isRTP := true
			if (channel % 2) != 0 {
				channel--
				isRTP = false
			}

			// forward frame only if it has been set up
			if trackID, ok := sc.session.tcpTracksByChannel[channel]; ok {
				err := processFunc(trackID, isRTP, frame.Payload)
				if err != nil {
					return err
				}
			}

		case *base.Request:
			reqErr := make(chan error)
			select {
			case readRequest <- reqWrapper{req: &req, res: reqErr}:
				err := <-reqErr
				if err != nil {
					return err
				}

			case <-sc.ctx.Done():
				return liberrors.ErrorServerTerminated()
			}
		}
	}
}

func ErrorPayloadSizeGreaterThanMaximum(l, m int) error {
	return fmt.Errorf("payload size (%d) greater than maximum allowed (%d)", l, m)
}

func (sc *ServerCxn) handleRequest(req *base.Request) (*base.Response, error) {
	if cSeq, ok := req.Header[base.HeaderCSeq]; !ok || len(cSeq) != 1 {
		return &base.Response{
			StatusCode: base.StatusBadRequest,
			Header:     base.Header{},
		}, liberrors.ErrorCSeqMissing()
	}

	sxID := getSessionID(req.Header)

	switch req.Method {
	case base.Options:
		if sxID != "" {
			return sc.handleRequestInSession(sxID, req, false)
		}

		// handle request here
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
				base.HeaderPublic: base.HeaderValue{strings.Join(methods, ", ")},
			},
		}, nil

	case base.Describe:
		if h, ok := sc.s.Handler.(ServerHandlerOnDescribe); ok {
			pathAndQuery, ok := req.URL.RTSPPathAndQuery()
			if !ok {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorInvalidPath()
			}

			path, query := base.PathSplitQuery(pathAndQuery)

			res, stream, err := h.OnDescribe(&ServerHandlerOnDescribeCtx{
				Conn:    sc,
				Request: req,
				Path:    path,
				Query:   query,
			})

			if res.StatusCode == base.StatusOK {
				if res.Header == nil {
					res.Header = make(base.Header)
				}

				res.Header[base.HeaderContentBase] = base.HeaderValue{req.URL.String() + "/"}
				res.Header[base.HeaderContentType] = base.HeaderValue{"application/sdp"}

				// VLC uses multicast if the SDP contains a multicast address.
				// therefore, we introduce a special Query (vlcmulticast) that allows
				// to return a SDP that contains a multicast address.
				multicast := false
				if sc.s.MulticastIPRange != "" {
					if q, err := url.ParseQuery(query); err == nil {
						if _, ok := q[queryVLCMulticast]; ok {
							multicast = true
						}
					}
				}

				if stream != nil {
					res.Body = stream.Tracks().Write(multicast)
				}
			}

			return res, err
		}

	case base.Announce:
		if _, ok := sc.s.Handler.(ServerHandlerOnAnnounce); ok {
			return sc.handleRequestInSession(sxID, req, true)
		}

	case base.Setup:
		if _, ok := sc.s.Handler.(ServerHandlerOnSetup); ok {
			return sc.handleRequestInSession(sxID, req, true)
		}

	case base.Play:
		if sxID != "" {
			if _, ok := sc.s.Handler.(ServerHandlerOnPlay); ok {
				return sc.handleRequestInSession(sxID, req, false)
			}
		}

	case base.Record:
		if sxID != "" {
			if _, ok := sc.s.Handler.(ServerHandlerOnRecord); ok {
				return sc.handleRequestInSession(sxID, req, false)
			}
		}

	case base.Pause:
		if sxID != "" {
			if _, ok := sc.s.Handler.(ServerHandlerOnPause); ok {
				return sc.handleRequestInSession(sxID, req, false)
			}
		}

	case base.Teardown:
		if sxID != "" {
			return sc.handleRequestInSession(sxID, req, false)
		}

	case base.GetParameter:
		if sxID != "" {
			return sc.handleRequestInSession(sxID, req, false)
		}

		// handle request here
		if h, ok := sc.s.Handler.(ServerHandlerOnGetParameter); ok {
			pathAndQuery, ok := req.URL.RTSPPathAndQuery()
			if !ok {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorInvalidPath()
			}

			path, query := base.PathSplitQuery(pathAndQuery)

			return h.OnGetParameter(&ServerHandlerOnGetParameterCtx{
				Conn:    sc,
				Request: req,
				Path:    path,
				Query:   query,
			})
		}

	case base.SetParameter:
		if h, ok := sc.s.Handler.(ServerHandlerOnSetParameter); ok {
			pathAndQuery, ok := req.URL.RTSPPathAndQuery()
			if !ok {
				return &base.Response{
					StatusCode: base.StatusBadRequest,
				}, liberrors.ErrorInvalidPath()
			}

			path, query := base.PathSplitQuery(pathAndQuery)

			return h.OnSetParameter(&ServerHandlerOnSetParameterCtx{
				Conn:    sc,
				Request: req,
				Path:    path,
				Query:   query,
			})
		}
	}

	return &base.Response{
		StatusCode: base.StatusBadRequest,
	}, liberrors.ErrorUnhandledRequest(req.Method.String(), req.URL.String())
}

func (sc *ServerCxn) handleRequestOuter(req *base.Request) error {
	if h, ok := sc.s.Handler.(ServerHandlerOnRequest); ok {
		h.OnRequest(sc, req)
	}

	res, err := sc.handleRequest(req)
	if res.Header == nil {
		res.Header = make(base.Header)
	}

	// add cSeq
	if _, ok := err.(liberrors.ErrCSeqMissing); !ok {
		res.Header[base.HeaderCSeq] = req.Header[base.HeaderCSeq]
	}

	// add server
	res.Header[base.HeaderServer] = base.HeaderValue{base.SeverName}
	if h, ok := sc.s.Handler.(ServerHandlerOnResponse); ok {
		h.OnResponse(sc, res)
	}

	var buf bytes.Buffer
	res.Write(&buf)

	sc.cxn.SetWriteDeadline(time.Now().Add(sc.s.WriteTimeout))
	sc.cxn.Write(buf.Bytes())

	return err
}

func (sc *ServerCxn) handleRequestInSession(id string, req *base.Request, create bool) (badReq *base.Response, err error) {
	badReq = &base.Response{
		StatusCode: base.StatusBadRequest,
	}
	err = liberrors.ErrorServerTerminated()

	// handle directly in Session
	if sc.session != nil {
		// session ID is optional in SETUP and ANNOUNCE requests, since
		// client may not have received the session ID yet due to multiple reasons:
		// * requests can be retries after code 301
		// * SETUP requests comes after ANNOUNCE response, that don't contain the session ID
		if id != "" {
			// the connection can't communicate with two sessions at once.
			if id != sc.session.secretID {
				return badReq, liberrors.ErrorCxnLinkedToOtherSession()
			}
		}

		cRes := make(chan sessionRequestRes)
		sReq := sessionRequestReq{
			sc:     sc,
			req:    req,
			id:     id,
			create: create,
			res:    cRes,
		}

		select {
		case sc.session.request <- sReq:
			res := <-cRes
			sc.session = res.ss
			return res.res, res.err

		case <-sc.session.ctx.Done():
			return
		}
	}

	// otherwise, pass through Server
	cRes := make(chan sessionRequestRes)
	sReq := sessionRequestReq{
		sc:     sc,
		req:    req,
		id:     id,
		create: create,
		res:    cRes,
	}

	select {
	case sc.s.sessionRequest <- sReq:
		res := <-cRes
		sc.session = res.ss
		return res.res, res.err

	case <-sc.s.ctx.Done():
		return
	}
}
