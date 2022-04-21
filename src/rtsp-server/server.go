// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-05
package rtsp_server

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/base"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
)

type sessionRequestRes struct {
	ss  *ServerSession
	res *base.Response
	err error
}

type sessionRequestReq struct {
	sc     *ServerCxn
	req    *base.Request
	id     string
	create bool
	res    chan sessionRequestRes
}

type streamMulticastIPReq struct {
	res chan net.IP
}

// Server is a RTSP server.
type Server struct {
	//
	// handler
	//
	// an handler to handle server events.
	Handler ServerHandler

	//
	// RTSP parameters
	//
	// timeout of read operations.
	// It defaults to 10 seconds
	ReadTimeout time.Duration
	// timeout of write operations.
	// It defaults to 10 seconds
	WriteTimeout time.Duration
	// the RTSP address of the server, to accept connections and send and receive
	// packets with the TCP transport.
	RTSPAddress string
	// a TLS configuration to accept TLS (RTSPS) connections.
	TLSConfig *tls.Config
	// a port to send and receive RTP packets with the UDP transport.
	// If UDPRTPAddress and UDPRTCPAddress are filled, the server can support the UDP transport.
	UDPRTPAddress string
	// a port to send and receive RTCP packets with the UDP transport.
	// If UDPRTPAddress and UDPRTCPAddress are filled, the server can support the UDP transport.
	UDPRTCPAddress string
	// a range of multicast IPs to use with the UDP-multicast transport.
	// If MulticastIPRange, MulticastRTPPort, MulticastRTCPPort are filled, the server
	// can support the UDP-multicast transport.
	MulticastIPRange string
	// a port to send RTP packets with the UDP-multicast transport.
	// If MulticastIPRange, MulticastRTPPort, MulticastRTCPPort are filled, the server
	// can support the UDP-multicast transport.
	MulticastRTPPort int
	// a port to send RTCP packets with the UDP-multicast transport.
	// If MulticastIPRange, MulticastRTPPort, MulticastRTCPPort are filled, the server
	// can support the UDP-multicast transport.
	MulticastRTCPPort int
	// read buffer count.
	// If greater than 1, allows passing buffers to other routines different to the one
	// that is reading frames.
	// It also allows buffering routed frames and mitigate network fluctuations
	// that are particularly relevant when using UDP.
	// It defaults to 256.
	ReadBufferCount int
	// write buffer count.
	// It allows queuing packets before sending them.
	// It defaults to 256.
	WriteBufferCount int

	// System Functions

	// Listen function used to initialize the TCP listener.
	// It defaults to net.Listen.
	Listen func(network string, address string) (net.Listener, error)
	// ListenPacket function used to initialize UDP listeners.
	// It defaults to net.ListenPacket.
	ListenPacket func(network, address string) (net.PacketConn, error)

	udpReceiverReportPeriod time.Duration
	udpSenderReportPeriod   time.Duration
	sessionTimeout          time.Duration
	checkStreamPeriod       time.Duration

	ctx             context.Context
	ctxCancel       func()
	wg              sync.WaitGroup
	multicastNet    *net.IPNet
	multicastNextIP net.IP
	tcpListener     net.Listener
	udpRTPListener  *serverUDPListener
	udpRTCPListener *serverUDPListener
	sessions        map[string]*ServerSession
	cxns            map[*ServerCxn]struct{}
	closeError      error

	// in
	connClose         chan *ServerCxn
	sessionRequest    chan sessionRequestReq
	sessionClose      chan *ServerSession
	streamMulticastIP chan streamMulticastIPReq
}

// Start starts the server.
func (s *Server) Start() error {
	// RTSP parameters
	if s.ReadTimeout == 0 {
		s.ReadTimeout = defaultTimeout
	}
	if s.WriteTimeout == 0 {
		s.WriteTimeout = defaultTimeout
	}
	if s.ReadBufferCount == 0 {
		s.ReadBufferCount = defaultBufferCount
	}
	if s.WriteBufferCount == 0 {
		s.WriteBufferCount = defaultBufferCount
	}

	// system functions
	if s.Listen == nil {
		s.Listen = net.Listen
	}
	if s.ListenPacket == nil {
		s.ListenPacket = net.ListenPacket
	}

	// private
	if s.udpReceiverReportPeriod == 0 {
		s.udpReceiverReportPeriod = defaultTimeout
	}
	if s.udpSenderReportPeriod == 0 {
		s.udpSenderReportPeriod = defaultTimeout
	}
	if s.sessionTimeout == 0 {
		s.sessionTimeout = defaultSessionTimeout
	}
	if s.checkStreamPeriod == 0 {
		s.checkStreamPeriod = defaultStreamPeriod
	}

	if s.TLSConfig != nil && s.UDPRTPAddress != "" {
		return liberrors.ErrorTLSCantBeUsedWithUDP()
	}

	if s.TLSConfig != nil && s.MulticastIPRange != "" {
		return liberrors.ErrorTLSCantBeUsedWithUDPMulticast()
	}

	if s.RTSPAddress == "" {
		return liberrors.ErrorRTSPAddressNotProvided()
	}

	if (s.UDPRTPAddress != "" && s.UDPRTCPAddress == "") ||
		(s.UDPRTPAddress == "" && s.UDPRTCPAddress != "") {
		return liberrors.ErrorUDPAddressesMustBeUsedTogether()
	}

	if s.UDPRTPAddress != "" {
		rtpPort, err := extractPort(s.UDPRTPAddress)
		if err != nil {
			return err
		}

		rtcpPort, err := extractPort(s.UDPRTCPAddress)
		if err != nil {
			return err
		}

		if (rtpPort % 2) != 0 {
			return liberrors.ErrorRTPPortMustBeEven()
		}

		if rtcpPort != (rtpPort + 1) {
			return liberrors.ErrorRTPPortsMustBeConsecutive()
		}

		s.udpRTPListener, err = newServerUDPListener(s, false, s.UDPRTPAddress, true)
		if err != nil {
			return err
		}

		s.udpRTCPListener, err = newServerUDPListener(s, false, s.UDPRTCPAddress, false)
		if err != nil {
			s.udpRTPListener.close()
			return err
		}
	}

	if s.MulticastIPRange != "" && (s.MulticastRTPPort == 0 || s.MulticastRTCPPort == 0) ||
		(s.MulticastRTPPort != 0 && (s.MulticastRTCPPort == 0 || s.MulticastIPRange == "")) ||
		s.MulticastRTCPPort != 0 && (s.MulticastRTPPort == 0 || s.MulticastIPRange == "") {
		if s.udpRTPListener != nil {
			s.udpRTPListener.close()
		}
		if s.udpRTCPListener != nil {
			s.udpRTCPListener.close()
		}
		
		return liberrors.ErrorMulticastInfoMustBeUsedTogether()
	}

	if s.MulticastIPRange != "" {
		if (s.MulticastRTPPort % 2) != 0 {
			if s.udpRTPListener != nil {
				s.udpRTPListener.close()
			}

			if s.udpRTCPListener != nil {
				s.udpRTCPListener.close()
			}

			return liberrors.ErrorRTPPortMustBeEven()
		}

		if s.MulticastRTCPPort != (s.MulticastRTPPort + 1) {
			if s.udpRTPListener != nil {
				s.udpRTPListener.close()
			}

			if s.udpRTCPListener != nil {
				s.udpRTCPListener.close()
			}

			return liberrors.ErrorRTPPortsMustBeConsecutive()
		}

		var err error
		_, s.multicastNet, err = net.ParseCIDR(s.MulticastIPRange)
		if err != nil {
			if s.udpRTPListener != nil {
				s.udpRTPListener.close()
			}

			if s.udpRTCPListener != nil {
				s.udpRTCPListener.close()
			}

			return err
		}

		s.multicastNextIP = s.multicastNet.IP
	}

	var err error
	s.tcpListener, err = s.Listen("tcp", s.RTSPAddress)
	if err != nil {
		if s.udpRTPListener != nil {
			s.udpRTPListener.close()
		}
		if s.udpRTCPListener != nil {
			s.udpRTCPListener.close()
		}
		return err
	}

	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	s.wg.Add(1)
	go s.run()

	return nil
}

// Close closes all the server resources and waits for them to close.
func (s *Server) Close() error {
	s.ctxCancel()
	s.wg.Wait()
	return s.closeError
}

// Wait waits until all server resources are closed.
// This can happen when a fatal error occurs or when Close() is called.
func (s *Server) Wait() error {
	s.wg.Wait()
	return s.closeError
}

func (s *Server) run() {
	defer s.wg.Done()

	s.sessions = make(map[string]*ServerSession)
	s.cxns = make(map[*ServerCxn]struct{})
	s.connClose = make(chan *ServerCxn)
	s.sessionRequest = make(chan sessionRequestReq)
	s.sessionClose = make(chan *ServerSession)
	s.streamMulticastIP = make(chan streamMulticastIPReq)

	s.wg.Add(1)
	cxns := make(chan net.Conn)
	acceptErr := make(chan error)
	go func() {
		defer s.wg.Done()
		err := func() error {
			for {
				cxn, err := s.tcpListener.Accept()
				if err != nil {
					return err
				}

				select {
				case cxns <- cxn:
				case <-s.ctx.Done():
					cxn.Close()
				}
			}
		}()

		select {
		case acceptErr <- err:
		case <-s.ctx.Done():
		}
	}()

	s.closeError = func() error {
		for {
			select {
			case err := <-acceptErr:
				return err

			case cxn := <-cxns:
				sc := newServerCxn(s, cxn)
				s.cxns[sc] = struct{}{}

			case sc := <-s.connClose:
				if _, ok := s.cxns[sc]; !ok {
					continue
				}
				delete(s.cxns, sc)
				sc.Close()

			case req := <-s.sessionRequest:
				if ss, ok := s.sessions[req.id]; ok {
					if !req.sc.ip().Equal(ss.author.ip()) ||
						req.sc.zone() != ss.author.zone() {
						req.res <- sessionRequestRes{
							res: &base.Response{
								StatusCode: base.StatusBadRequest,
							},
							err: liberrors.ErrorCannotUseSessionCreatedByOtherIP(),
						}
						continue
					}

					ss.request <- req
				} else {
					if !req.create {
						req.res <- sessionRequestRes{
							res: &base.Response{
								StatusCode: base.StatusSessionNotFound,
							},
							err: liberrors.ErrorSessionNotFound(),
						}
						continue
					}

					secretID, err := newSessionSecretID(s.sessions)
					if err != nil {
						req.res <- sessionRequestRes{
							res: &base.Response{
								StatusCode: base.StatusBadRequest,
							},
							err: fmt.Errorf("internal error"),
						}
						continue
					}

					ss := newServerSession(s, secretID, req.sc)
					s.sessions[secretID] = ss

					select {
					case ss.request <- req:
					case <-ss.ctx.Done():
						req.res <- sessionRequestRes{
							res: &base.Response{
								StatusCode: base.StatusBadRequest,
							},
							err: liberrors.ErrorServerTerminated(),
						}
					}
				}

			case ss := <-s.sessionClose:
				if sss, ok := s.sessions[ss.secretID]; !ok || sss != ss {
					continue
				}
				delete(s.sessions, ss.secretID)
				ss.Close()

			case req := <-s.streamMulticastIP:
				ip32 := binary.BigEndian.Uint32(s.multicastNextIP)
				mask := binary.BigEndian.Uint32(s.multicastNet.Mask)
				ip32 = (ip32 & mask) | ((ip32 + 1) & ^mask)
				ip := make(net.IP, 4)
				binary.BigEndian.PutUint32(ip, ip32)
				s.multicastNextIP = ip
				req.res <- ip

			case <-s.ctx.Done():
				return liberrors.ErrorServerTerminated()
			}
		}
	}()

	s.ctxCancel()

	if s.udpRTCPListener != nil {
		s.udpRTCPListener.close()
	}

	if s.udpRTPListener != nil {
		s.udpRTPListener.close()
	}

	s.tcpListener.Close()
}

// StartAndWait starts the server and waits until a fatal error.
func (s *Server) StartAndWait() error {
	err := s.Start()
	if err != nil {
		return err
	}

	return s.Wait()
}
