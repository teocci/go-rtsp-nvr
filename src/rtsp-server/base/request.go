// Package base
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-06
package base

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
	"strconv"
)

// Request is a RTSP request.
type Request struct {
	// request method
	Method Method

	// request url
	URL *URL

	// map of header values
	Header Header

	// optional body
	Body []byte
}

// Read reads a request.
func (req *Request) Read(rb *bufio.Reader) error {
	bts, err := readBytesLimited(rb, ' ', requestMaxMethodLength)
	if err != nil {
		return err
	}
	req.Method = Method(bts[:len(bts)-1])

	if req.Method == "" {
		return liberrors.ErrorEmptyMethod()
	}

	bts, err = readBytesLimited(rb, ' ', requestMaxURLLength)
	if err != nil {
		return err
	}
	rawURL := string(bts[:len(bts)-1])

	ur, err := ParseURL(rawURL)
	if err != nil {
		return ErrorInvalidURL(rawURL)
	}
	req.URL = ur

	bts, err = readBytesLimited(rb, '\r', requestMaxProtocolLength)
	if err != nil {
		return err
	}

	proto := bts[:len(bts)-1]
	if string(proto) != rtspProtocol10 {
		return liberrors.ErrorProtocolVersionNotExpected(rtspProtocol10, proto)
	}

	err = readByteEqual(rb, '\n')
	if err != nil {
		return err
	}

	err = req.Header.read(rb)
	if err != nil {
		return err
	}

	err = (*body)(&req.Body).read(req.Header, rb)
	if err != nil {
		return err
	}

	return nil
}

func ErrorInvalidURL(url string) error {
	return fmt.Errorf("invalid URL (%v)", url)
}

// ReadIgnoreFrames reads a request and ignores any interleaved frame sent
// before the request.
func (req *Request) ReadIgnoreFrames(maxPayloadSize int, rb *bufio.Reader) error {
	var f InterleavedFrame

	for {
		buff, err := ReadInterleavedFrameOrRequest(&f, maxPayloadSize, req, rb)
		if err != nil {
			return err
		}

		if _, ok := buff.(*Request); ok {
			return nil
		}
	}
}

// Write writes a request.
func (req Request) Write(bb *bytes.Buffer) {
	bb.Reset()

	urStr := req.URL.CloneWithoutCredentials().String()
	bb.Write([]byte(string(req.Method) + " " + urStr + " " + rtspProtocol10 + "\r\n"))

	if len(req.Body) != 0 {
		req.Header["Content-Length"] = HeaderValue{strconv.FormatInt(int64(len(req.Body)), 10)}
	}

	req.Header.write(bb)

	body(req.Body).write(bb)
}

// String implements fmt.Stringer.
func (req Request) String() string {
	var buf bytes.Buffer
	req.Write(&buf)
	return buf.String()
}
