// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-28
package webserver

import (
	"github.com/teocci/go-stream-av/av"
)

func SupportedCodecs() []av.CodecType {
	return []av.CodecType{
		av.H264,
		av.PCM_ALAW,
		av.PCM_MULAW,
		av.OPUS,
	}
}

func isSupported(c av.CodecType) bool {
	for _, codecType := range SupportedCodecs() {
		if c == codecType {
			return true
		}
	}

	return false
}
