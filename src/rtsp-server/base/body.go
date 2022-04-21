// Package base
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-07
package base

import (
	"bufio"
	"bytes"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
	"io"
	"strconv"
)

type body []byte

func (b *body) read(header Header, rb *bufio.Reader) error {
	headerValue, ok := header["Content-Length"]
	if !ok || len(headerValue) != 1 {
		*b = nil
		return nil
	}

	cl, err := strconv.ParseInt(headerValue[0], 10, 64)
	if err != nil {
		return liberrors.ErrorInvalidContentLength()
	}

	if cl > rtspMaxContentLength {
		return liberrors.ErrorMaxContentLengthExceeded(rtspMaxContentLength, cl)
	}

	*b = make([]byte, cl)
	n, err := io.ReadFull(rb, *b)
	if err != nil && n != len(*b) {
		return err
	}

	return nil
}

func (b body) write(bb *bytes.Buffer) {
	if len(b) == 0 {
		return
	}

	bb.Write(b)
}
