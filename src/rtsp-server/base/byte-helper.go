// Package base
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-06
package base

import (
	"bufio"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
)

func readByteEqual(rb *bufio.Reader, cmp byte) error {
	byt, err := rb.ReadByte()
	if err != nil {
		return err
	}

	if byt != cmp {
		return liberrors.ErrorCharNotExpected(cmp, byt)
	}

	return nil
}

func readBytesLimited(rb *bufio.Reader, delim byte, n int) ([]byte, error) {
	for i := 1; i <= n; i++ {
		bts, err := rb.Peek(i)
		if err != nil {
			return nil, err
		}

		if bts[len(bts)-1] == delim {
			rb.Discard(len(bts))
			return bts, nil
		}
	}

	return nil, liberrors.ErrorBufferLengthExceeds(n)
}
