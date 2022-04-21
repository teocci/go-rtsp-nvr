// Package rtcpreceiver
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-07
package rtcpreceiver

import (
	"crypto/rand"
	"time"
)

func randUint32() uint32 {
	var b [4]byte
	rand.Read(b[:])
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

var now = time.Now
