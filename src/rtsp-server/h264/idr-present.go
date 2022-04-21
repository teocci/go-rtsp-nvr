// Package h264
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-13
package h264

// IDRPresent check if there's an IDR inside provided NALUs.
func IDRPresent(nalus [][]byte) bool {
	for _, n := range nalus {
		idr := NALUType(n[0] & 0x1F)
		if idr == NALUTypeIDR {
			return true
		}
	}

	return false
}
