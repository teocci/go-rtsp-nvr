// Package h264
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-13
package h264

import (
	"encoding/binary"
	"fmt"
)

// DecodeAVCC decodes NALUs from the AVCC stream format.
func DecodeAVCC(bts []byte) ([][]byte, error) {
	var ret [][]byte

	for len(bts) > 0 {
		if len(bts) < 4 {
			return nil, fmt.Errorf("invalid length")
		}

		le := binary.BigEndian.Uint32(bts)
		bts = bts[4:]

		if len(bts) < int(le) {
			return nil, fmt.Errorf("invalid length")
		}

		ret = append(ret, bts[:le])
		bts = bts[le:]
	}

	if len(ret) == 0 {
		return nil, fmt.Errorf("no NALUs decoded")
	}

	return ret, nil
}

// EncodeAVCC encodes NALUs into the AVCC stream format.
func EncodeAVCC(nalus [][]byte) ([]byte, error) {
	le := 0
	for _, nalu := range nalus {
		le += 4 + len(nalu)
	}

	ret := make([]byte, le)
	pos := 0

	for _, nalu := range nalus {
		ln := len(nalu)
		binary.BigEndian.PutUint32(ret[pos:], uint32(ln))
		pos += 4

		copy(ret[pos:], nalu)
		pos += ln
	}

	return ret, nil
}
