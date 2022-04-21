// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-08
package rtsp_server

import "time"

const (
	tcpReadBufferSize             = 4096
	serverUDPKernelReadBufferSize = 0x80000 // same as gstreamer's rtsp-src

	// this must fit an entire H264 NALU and a RTP header.
	// with a 250 Mbps H264 video, the maximum NALU size is 2.2MB
	tcpMaxFramePayloadSize = 3 * 1024 * 1024

	defaultStreamPeriod    = 1 * time.Second
	defaultTimeout         = 10 * time.Second
	defaultKeepalivePeriod = 30 * time.Second
	defaultSessionTimeout  = 1 * 60 * time.Second

	defaultBufferCount    = 256
	defaultReadBufferSize = 2048

	queryVLCMulticast = "vlcmulticast"

	keyParameterSets = "sprop-parameter-sets"

	mediaDescriptorRTPMapKey = "rtpmap"

	// same size as GStreamer's rtspSRC
	udpKernelReadBufferSize = 0x80000

	// same size as GStreamer's rtspSRC
	multicastTTL = 16

	// 1500 (UDP MTU) - 20 (IP header) - 8 (UDP header)
	maxPacketSize = 1472
)
