// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-08
package rtsp_server

import (
	"github.com/pion/rtp"
)

type rtpPacketMultiBuffer struct {
	count   uint64
	buffers []rtp.Packet
	cur     uint64
}

func newRTPPacketMultiBuffer(count uint64) *rtpPacketMultiBuffer {
	return &rtpPacketMultiBuffer{
		count:   count,
		buffers: make([]rtp.Packet, count),
	}
}

func (mb *rtpPacketMultiBuffer) next() *rtp.Packet {
	ret := &mb.buffers[mb.cur%mb.count]
	mb.cur++
	return ret
}
