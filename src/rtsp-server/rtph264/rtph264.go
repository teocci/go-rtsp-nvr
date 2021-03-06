// Package rtph264
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-13
package rtph264

const (
	rtpVersion   = 0x02
	rtpClockRate = 90000 // h264 always uses 90khz

	// with a 250 Mbps H264 video, the maximum NALU size is 2.2MB
	maxNALUSize = 3 * 1024 * 1024
)
