// Package h264
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-13
package h264

import (
	"fmt"
)

// DecodeAnnexB decodes NALUs from the Annex-B stream format.
func DecodeAnnexB(bts []byte) ([][]byte, error) {
	bl := len(bts)
	zeroCount := 0

outer:
	for i := 0; i < bl; i++ {
		switch bts[i] {
		case 0:
			zeroCount++

		case 1:
			break outer

		default:
			return nil, fmt.Errorf("unexpected byte: %d", bts[i])
		}
	}
	if zeroCount != 2 && zeroCount != 3 {
		return nil, fmt.Errorf("initial delimiter not found")
	}

	var ret [][]byte
	start := zeroCount + 1
	zeroCount = 0
	delimStart := 0

	for i := start; i < bl; i++ {
		switch bts[i] {
		case 0:
			if zeroCount == 0 {
				delimStart = i
			}
			zeroCount++

		case 1:
			if zeroCount == 2 || zeroCount == 3 {
				nalu := bts[start:delimStart]
				if len(nalu) == 0 {
					return nil, fmt.Errorf("empty NALU")
				}

				ret = append(ret, nalu)
				start = i + 1
			}
			zeroCount = 0

		default:
			zeroCount = 0
		}
	}

	nalu := bts[start:bl]
	if len(nalu) == 0 {
		return nil, fmt.Errorf("empty NALU")
	}
	ret = append(ret, nalu)

	return ret, nil
}

// EncodeAnnexB encodes NALUs into the Annex-B stream format.
func EncodeAnnexB(nalus [][]byte) ([]byte, error) {
	var ret []byte

	for _, nalu := range nalus {
		ret = append(ret, []byte{0x00, 0x00, 0x00, 0x01}...)
		ret = append(ret, nalu...)
	}

	return ret, nil
}
