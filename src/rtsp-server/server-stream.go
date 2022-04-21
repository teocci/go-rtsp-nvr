// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-08
package rtsp_server

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"

	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/rtcpsender"
)

type serverStreamTrack struct {
	padding            uint32 //nolint:structcheck,unused
	lastSequenceNumber uint32
	lastSSRC           uint32
	lastTimeRTP        uint32
	lastTimeNTP        int64
	rtcpSender         *rtcpsender.RTCPSender
}

// ServerStream represents a single stream.
// This is in charge of
// - distributing the stream to each reader
// - allocating multicast listeners
// - gathering infos about the stream to generate SSRC and RTP-Info
type ServerStream struct {
	tracks Tracks

	mutex                   sync.RWMutex
	s                       *Server
	readersUnicast          map[*ServerSession]struct{}
	readers                 map[*ServerSession]struct{}
	serverMulticastHandlers []*serverMulticastHandler
	stTracks                []*serverStreamTrack
}

// NewServerStream allocates a ServerStream.
func NewServerStream(tracks Tracks) *ServerStream {
	tracks = tracks.clone()
	tracks.setControls()

	st := &ServerStream{
		tracks:         tracks,
		readersUnicast: make(map[*ServerSession]struct{}),
		readers:        make(map[*ServerSession]struct{}),
	}

	st.stTracks = make([]*serverStreamTrack, len(tracks))
	for i := range st.stTracks {
		st.stTracks[i] = &serverStreamTrack{}
	}

	return st
}

// Close closes a ServerStream.
func (st *ServerStream) Close() error {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	for ss := range st.readers {
		ss.Close()
	}

	if st.serverMulticastHandlers != nil {
		for _, h := range st.serverMulticastHandlers {
			h.close()
		}
		st.serverMulticastHandlers = nil
	}

	st.readers = nil
	st.readersUnicast = nil

	return nil
}

// Tracks returns the tracks of the stream.
func (st *ServerStream) Tracks() Tracks {
	return st.tracks
}

func (st *ServerStream) ssrc(trackID int) uint32 {
	return atomic.LoadUint32(&st.stTracks[trackID].lastSSRC)
}

func (st *ServerStream) timestamp(trackID int) uint32 {
	lastTimeRTP := atomic.LoadUint32(&st.stTracks[trackID].lastTimeRTP)
	lastTimeNTP := atomic.LoadInt64(&st.stTracks[trackID].lastTimeNTP)

	if lastTimeRTP == 0 || lastTimeNTP == 0 {
		return 0
	}

	return uint32(uint64(lastTimeRTP) +
		uint64(time.Since(time.Unix(lastTimeNTP, 0)).Seconds()*float64(st.tracks[trackID].ClockRate())))
}

func (st *ServerStream) lastSequenceNumber(trackID int) uint16 {
	return uint16(atomic.LoadUint32(&st.stTracks[trackID].lastSequenceNumber))
}

func (st *ServerStream) readerAdd(ss *ServerSession, transport Transport, clientPorts *[2]int) error {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if st.s == nil {
		st.s = ss.s

		for trackID, track := range st.stTracks {
			cTrackID := trackID
			track.rtcpSender = rtcpsender.New(
				st.s.udpSenderReportPeriod,
				st.tracks[trackID].ClockRate(),
				func(pkt rtcp.Packet) {
					st.writePacketRTCPSenderReport(cTrackID, pkt)
				},
			)
		}
	}

	switch transport {
	case TransportUDP:
		// check if client ports are already in use by another reader.
		for r := range st.readersUnicast {
			if *r.setupTransport == TransportUDP &&
				r.author.ip().Equal(ss.author.ip()) &&
				r.author.zone() == ss.author.zone() {
				for _, rt := range r.setupTracks {
					if rt.udpRTPPort == clientPorts[0] {
						return liberrors.ErrorServerUDPPortsAlreadyInUse(rt.udpRTPPort)
					}
				}
			}
		}

	case TransportUDPMulticast:
		// allocate multicast listeners
		if st.serverMulticastHandlers == nil {
			st.serverMulticastHandlers = make([]*serverMulticastHandler, len(st.tracks))

			for i := range st.tracks {
				h, err := newServerMulticastHandler(st.s)
				if err != nil {
					for _, h := range st.serverMulticastHandlers {
						if h != nil {
							h.close()
						}
					}
					st.serverMulticastHandlers = nil
					return err
				}

				st.serverMulticastHandlers[i] = h
			}
		}
	}

	st.readers[ss] = struct{}{}

	return nil
}

func (st *ServerStream) readerRemove(ss *ServerSession) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	delete(st.readers, ss)

	if len(st.readers) == 0 && st.serverMulticastHandlers != nil {
		for _, l := range st.serverMulticastHandlers {
			l.rtpListener.close()
			l.rtcpListener.close()
		}
		st.serverMulticastHandlers = nil
	}
}

func (st *ServerStream) readerSetActive(ss *ServerSession) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	switch *ss.setupTransport {
	case TransportUDP, TransportTCP:
		st.readersUnicast[ss] = struct{}{}

	default: // UDPMulticast
		for trackID := range ss.setupTracks {
			st.serverMulticastHandlers[trackID].rtcpListener.addClient(
				ss.author.ip(), st.serverMulticastHandlers[trackID].rtcpListener.port(), ss, trackID, false)
		}
	}
}

func (st *ServerStream) readerSetInactive(ss *ServerSession) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	switch *ss.setupTransport {
	case TransportUDP, TransportTCP:
		delete(st.readersUnicast, ss)

	default: // UDPMulticast
		if st.serverMulticastHandlers != nil {
			for trackID := range ss.setupTracks {
				st.serverMulticastHandlers[trackID].rtcpListener.removeClient(ss)
			}
		}
	}
}

// WritePacketRTP writes a RTP packet to all the readers of the stream.
func (st *ServerStream) WritePacketRTP(trackID int, pkt *rtp.Packet, ptsEqualsDTS bool) {
	byts := make([]byte, maxPacketSize)
	n, err := pkt.MarshalTo(byts)
	if err != nil {
		return
	}
	byts = byts[:n]

	track := st.stTracks[trackID]
	now := time.Now()

	atomic.StoreUint32(&track.lastSequenceNumber,
		uint32(pkt.Header.SequenceNumber))
	atomic.StoreUint32(&track.lastSSRC, pkt.Header.SSRC)

	if ptsEqualsDTS {
		atomic.StoreUint32(&track.lastTimeRTP, pkt.Header.Timestamp)
		atomic.StoreInt64(&track.lastTimeNTP, now.Unix())
	}

	st.mutex.RLock()
	defer st.mutex.RUnlock()

	if track.rtcpSender != nil {
		track.rtcpSender.ProcessPacketRTP(now, pkt, ptsEqualsDTS)
	}

	// send unicast
	for r := range st.readersUnicast {
		r.writePacketRTP(trackID, byts)
	}

	// send multicast
	if st.serverMulticastHandlers != nil {
		st.serverMulticastHandlers[trackID].writePacketRTP(byts)
	}
}

// WritePacketRTCP writes a RTCP packet to all the readers of the stream.
func (st *ServerStream) WritePacketRTCP(trackID int, pkt rtcp.Packet) {
	bts, err := pkt.Marshal()
	if err != nil {
		return
	}

	st.mutex.RLock()
	defer st.mutex.RUnlock()

	// send unicast
	for r := range st.readersUnicast {
		r.writePacketRTCP(trackID, bts)
	}

	// send multicast
	if st.serverMulticastHandlers != nil {
		st.serverMulticastHandlers[trackID].writePacketRTCP(bts)
	}
}

func (st *ServerStream) writePacketRTCPSenderReport(trackID int, pkt rtcp.Packet) {
	bts, err := pkt.Marshal()
	if err != nil {
		return
	}

	st.mutex.RLock()
	defer st.mutex.RUnlock()

	// send unicast (UDP only)
	for r := range st.readersUnicast {
		if *r.setupTransport == TransportUDP {
			r.writePacketRTCP(trackID, bts)
		}
	}

	// send multicast
	if st.serverMulticastHandlers != nil {
		st.serverMulticastHandlers[trackID].writePacketRTCP(bts)
	}
}
