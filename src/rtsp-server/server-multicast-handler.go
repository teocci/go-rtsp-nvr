// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-11
package rtsp_server

import (
	"net"

	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/ringbuffer"
)

type trackTypePayload struct {
	trackID int
	isRTP   bool
	payload []byte
}

type serverMulticastHandler struct {
	rtpListener  *serverUDPListener
	rtcpListener *serverUDPListener
	writeBuffer  *ringbuffer.RingBuffer

	writerDone chan struct{}
}

func newServerMulticastHandler(s *Server) (*serverMulticastHandler, error) {
	rtpl, rtcpl, err := newServerUDPListenerMulticastPair(s)
	if err != nil {
		return nil, err
	}

	h := &serverMulticastHandler{
		rtpListener:  rtpl,
		rtcpListener: rtcpl,
		writeBuffer:  ringbuffer.New(uint64(s.WriteBufferCount)),
		writerDone:   make(chan struct{}),
	}

	go h.runWriter()

	return h, nil
}

func (h *serverMulticastHandler) close() {
	h.rtpListener.close()
	h.rtcpListener.close()
	h.writeBuffer.Close()
	<-h.writerDone
}

func (h *serverMulticastHandler) ip() net.IP {
	return h.rtpListener.ip()
}

func (h *serverMulticastHandler) runWriter() {
	defer close(h.writerDone)

	rtpAddr := &net.UDPAddr{
		IP:   h.rtpListener.ip(),
		Port: h.rtpListener.port(),
	}

	rtcpAddr := &net.UDPAddr{
		IP:   h.rtcpListener.ip(),
		Port: h.rtcpListener.port(),
	}

	for {
		buffer, ok := h.writeBuffer.Pull()
		if !ok {
			return
		}

		data := buffer.(trackTypePayload)

		if data.isRTP {
			h.rtpListener.write(data.payload, rtpAddr)
		} else {
			h.rtcpListener.write(data.payload, rtcpAddr)
		}
	}
}

func (h *serverMulticastHandler) writePacketRTP(payload []byte) {
	h.writeBuffer.Push(trackTypePayload{
		isRTP:   true,
		payload: payload,
	})
}

func (h *serverMulticastHandler) writePacketRTCP(payload []byte) {
	h.writeBuffer.Push(trackTypePayload{
		isRTP:   false,
		payload: payload,
	})
}
