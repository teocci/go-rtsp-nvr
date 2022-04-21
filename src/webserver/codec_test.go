// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-28
package webserver

import (
	"testing"

	"github.com/teocci/go-stream-av/av"
)

func TestIsSupported(t *testing.T) {
	codecType := av.H264

	supported := isSupported(codecType)
	if !supported {
		t.Fatalf(`isSupported() = %t, want %t`, supported, true)
	}
}
