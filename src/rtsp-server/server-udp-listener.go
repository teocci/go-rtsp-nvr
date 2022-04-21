// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-08
package rtsp_server

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/rtcp"
	"golang.org/x/net/ipv4"

	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
)

type clientData struct {
	ss           *ServerSession
	trackID      int
	isPublishing bool
}

type clientAddr struct {
	ip   [net.IPv6len]byte // use a fixed-size array to enable the equality operator
	port int
}

func (p *clientAddr) fill(ip net.IP, port int) {
	p.port = port

	if len(ip) == net.IPv4len {
		copy(p.ip[0:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}) // v4InV6Prefix
		copy(p.ip[12:], ip)
	} else {
		copy(p.ip[:], ip)
	}
}

type serverUDPListener struct {
	s *Server

	pc              *net.UDPConn
	listenIP        net.IP
	isRTP           bool
	writeTimeout    time.Duration
	rtpPacketBuffer *rtpPacketMultiBuffer
	clientsMutex    sync.RWMutex
	clients         map[clientAddr]*clientData
	processFunc     func(*clientData, []byte)

	readerDone chan struct{}
}

func newServerUDPListenerMulticastPair(s *Server) (rtpListener *serverUDPListener, rtcpListener *serverUDPListener, err error) {
	res := make(chan net.IP)
	select {
	case s.streamMulticastIP <- streamMulticastIPReq{res: res}:
	case <-s.ctx.Done():
		return nil, nil, liberrors.ErrorServerTerminated()
	}

	ip := <-res

	address := mergeHostPortInt(ip.String(), s.MulticastRTPPort)
	rtpListener, err = newServerUDPListener(s, true, address, true)
	if err != nil {
		return nil, nil, err
	}

	address = mergeHostPortInt(ip.String(), s.MulticastRTCPPort)
	rtcpListener, err = newServerUDPListener(s, true, address, false)
	if err != nil {
		rtpListener.close()
		return nil, nil, err
	}

	return
}

func newServerUDPListener(s *Server, multicast bool, address string, isRTP bool) (listener *serverUDPListener, err error) {
	var udpPacket *net.UDPConn
	var listenIP net.IP
	var packet net.PacketConn
	var interfaces []net.Interface
	var host, port string

	if multicast {
		host, port, err = net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}

		packet, err = s.ListenPacket("udp", mergeHostPort("224.0.0.0", port))
		if err != nil {
			return nil, err
		}

		p := ipv4.NewPacketConn(packet)

		err = p.SetMulticastTTL(multicastTTL)
		if err != nil {
			return nil, err
		}

		interfaces, err = net.Interfaces()
		if err != nil {
			return nil, err
		}

		listenIP = net.ParseIP(host)

		for _, i := range interfaces {
			if (i.Flags & net.FlagMulticast) != 0 {
				err = p.JoinGroup(&i, &net.UDPAddr{IP: listenIP})
				if err != nil {
					return
				}
			}
		}

		udpPacket = packet.(*net.UDPConn)
	} else {
		packet, err = s.ListenPacket("udp", address)
		if err != nil {
			return
		}

		udpPacket = packet.(*net.UDPConn)
		listenIP = packet.LocalAddr().(*net.UDPAddr).IP
	}

	err = udpPacket.SetReadBuffer(udpKernelReadBufferSize)
	if err != nil {
		return
	}

	listener = &serverUDPListener{
		s:               s,
		pc:              udpPacket,
		listenIP:        listenIP,
		clients:         make(map[clientAddr]*clientData),
		isRTP:           isRTP,
		writeTimeout:    s.WriteTimeout,
		rtpPacketBuffer: newRTPPacketMultiBuffer(uint64(s.ReadBufferCount)),
		readerDone:      make(chan struct{}),
	}

	if isRTP {
		listener.processFunc = listener.processRTP
	} else {
		listener.processFunc = listener.processRTCP
	}

	go listener.runReader()

	return listener, nil
}

func (u *serverUDPListener) close() {
	_ = u.pc.Close()
	<-u.readerDone
}

func (u *serverUDPListener) ip() net.IP {
	return u.listenIP
}

func (u *serverUDPListener) port() int {
	return u.pc.LocalAddr().(*net.UDPAddr).Port
}

func (u *serverUDPListener) runReader() {
	defer close(u.readerDone)

	for {
		buf := make([]byte, maxPacketSize)
		n, addr, err := u.pc.ReadFromUDP(buf)
		if err != nil {
			break
		}

		func() {
			u.clientsMutex.RLock()
			defer u.clientsMutex.RUnlock()

			var cAddr clientAddr
			cAddr.fill(addr.IP, addr.Port)
			cData, ok := u.clients[cAddr]
			if !ok {
				return
			}

			u.processFunc(cData, buf[:n])
		}()
	}
}

func (u *serverUDPListener) processRTP(clientData *clientData, payload []byte) {
	pkt := u.rtpPacketBuffer.next()
	err := pkt.Unmarshal(payload)
	if err != nil {
		return
	}

	now := time.Now()
	atomic.StoreInt64(clientData.ss.udpLastFrameTime, now.Unix())

	ctx := ServerHandlerOnPacketRTPCtx{
		Session: clientData.ss,
		TrackID: clientData.trackID,
		Packet:  pkt,
	}
	at := clientData.ss.announcedTracks[clientData.trackID]
	clientData.ss.processPacketRTP(at, &ctx)

	at.rtcpReceiver.ProcessPacketRTP(now, ctx.Packet, ctx.PTSEqualsDTS)
	if h, ok := clientData.ss.s.Handler.(ServerHandlerOnPacketRTP); ok {
		h.OnPacketRTP(&ctx)
	}
}

func (u *serverUDPListener) processRTCP(clientData *clientData, payload []byte) {
	packets, err := rtcp.Unmarshal(payload)
	if err != nil {
		return
	}

	if clientData.isPublishing {
		now := time.Now()
		atomic.StoreInt64(clientData.ss.udpLastFrameTime, now.Unix())

		for _, pkt := range packets {
			clientData.ss.announcedTracks[clientData.trackID].rtcpReceiver.ProcessPacketRTCP(now, pkt)
		}
	}

	for _, pkt := range packets {
		clientData.ss.onPacketRTCP(clientData.trackID, pkt)
	}
}

func (u *serverUDPListener) write(buf []byte, addr *net.UDPAddr) error {
	// No mutex is needed here since Write() has an internal lock.
	// https://github.com/golang/go/issues/27203#issuecomment-534386117

	_ = u.pc.SetWriteDeadline(time.Now().Add(u.writeTimeout))
	_, err := u.pc.WriteTo(buf, addr)
	return err
}

func (u *serverUDPListener) addClient(ip net.IP, port int, ss *ServerSession, trackID int, isPublishing bool) {
	u.clientsMutex.Lock()
	defer u.clientsMutex.Unlock()

	var addr clientAddr
	addr.fill(ip, port)

	u.clients[addr] = &clientData{
		ss:           ss,
		trackID:      trackID,
		isPublishing: isPublishing,
	}
}

func (u *serverUDPListener) removeClient(ss *ServerSession) {
	u.clientsMutex.Lock()
	defer u.clientsMutex.Unlock()

	for addr, data := range u.clients {
		if data.ss == ss {
			delete(u.clients, addr)
		}
	}
}
