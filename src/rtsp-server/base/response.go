// Package base
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-06
package base

import (
	"bufio"
	"bytes"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
	"strconv"
)

// Response is a RTSP response.
type Response struct {
	// numeric status code
	StatusCode StatusCode

	// status message
	StatusMessage string

	// map of header values
	Header Header

	// optional body
	Body []byte
}

// Read reads a response.
func (res *Response) Read(rb *bufio.Reader) error {
	bts, err := readBytesLimited(rb, ' ', 255)
	if err != nil {
		return err
	}
	proto := bts[:len(bts)-1]

	if string(proto) != rtspProtocol10 {
		return liberrors.ErrorProtocolVersionNotExpected(rtspProtocol10, proto)
	}

	bts, err = readBytesLimited(rb, ' ', 4)
	if err != nil {
		return err
	}
	statusCodeStr := string(bts[:len(bts)-1])

	statusCode64, err := strconv.ParseInt(statusCodeStr, 10, 32)
	if err != nil {
		return liberrors.ErrorParsingStatusCode()
	}
	res.StatusCode = StatusCode(statusCode64)

	bts, err = readBytesLimited(rb, '\r', 255)
	if err != nil {
		return err
	}

	res.StatusMessage = string(bts[:len(bts)-1])
	if len(res.StatusMessage) == 0 {
		return liberrors.ErrorEmptyStatusMessage()
	}

	err = readByteEqual(rb, '\n')
	if err != nil {
		return err
	}

	err = res.Header.read(rb)
	if err != nil {
		return err
	}

	err = (*body)(&res.Body).read(res.Header, rb)
	if err != nil {
		return err
	}

	return nil
}

// ReadIgnoreFrames reads a response and ignores any interleaved frame sent
// before the response.
func (res *Response) ReadIgnoreFrames(maxPayloadSize int, rb *bufio.Reader) error {
	var f InterleavedFrame

	for {
		rec, err := ReadInterleavedFrameOrResponse(&f, maxPayloadSize, res, rb)
		if err != nil {
			return err
		}

		if _, ok := rec.(*Response); ok {
			return nil
		}
	}
}

// Write writes a Response.
func (res Response) Write(bb *bytes.Buffer) {
	bb.Reset()

	if res.StatusMessage == "" {
		if status, ok := statusMessages[res.StatusCode]; ok {
			res.StatusMessage = status
		}
	}

	bb.Write([]byte(rtspProtocol10 + " " + strconv.FormatInt(int64(res.StatusCode), 10) + " " + res.StatusMessage + "\r\n"))

	if len(res.Body) != 0 {
		res.Header["Content-Length"] = HeaderValue{strconv.FormatInt(int64(len(res.Body)), 10)}
	}

	res.Header.write(bb)

	body(res.Body).write(bb)
}

// String implements fmt.Stringer.
func (res Response) String() string {
	var buf bytes.Buffer
	res.Write(&buf)
	return buf.String()
}
