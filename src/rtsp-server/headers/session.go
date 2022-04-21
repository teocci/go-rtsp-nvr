// Package headers
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-07
package headers

import (
	"fmt"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
	"strconv"
	"strings"

	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/base"
)

const (
	formatSessionTimeout = ";timeout=%s"
)

// Session is a Session header.
type Session struct {
	// session id
	Session string

	// (optional) a timeout
	Timeout *uint
}

// Read decodes a Session header.
func (h *Session) Read(v base.HeaderValue) error {
	if len(v) == 0 {
		return liberrors.ErrorValueNotProvided()
	}

	if len(v) > 1 {
		return liberrors.ErrorValueProvidedMultipleTimes(v)
	}

	v0 := v[0]

	i := strings.IndexByte(v0, ';')
	if i < 0 {
		h.Session = v0
		return nil
	}

	h.Session = v0[:i]
	v0 = v0[i+1:]

	v0 = strings.TrimLeft(v0, " ")

	kvs, err := keyValParse(v0, ';')
	if err != nil {
		return err
	}

	for k, v := range kvs {
		if k == "timeout" {
			iv, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return err
			}

			uiv := uint(iv)
			h.Timeout = &uiv
		}
	}

	return nil
}

// Write encodes a Session header.
func (h Session) Write() base.HeaderValue {
	ret := h.Session

	if h.Timeout != nil {
		ret += fmt.Sprintf(formatSessionTimeout, strconv.FormatUint(uint64(*h.Timeout), 10))
	}

	return base.HeaderValue{ret}
}

