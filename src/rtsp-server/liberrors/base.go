// Package liberrors
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-28
package liberrors

import (
	"errors"
	"fmt"
)

const (
	errStreamNotFound             = "stream not found"
	errBadOption                  = "bad option: %s"
	errParsingOption              = "fail to parse option | buf: %s"
	errParsingStream              = "fail to parse stream | buf: %s"
	errParsingCSeq                = "fail to parse CSeq"
	errParsingStatusCode          = "unable to parse status code"
	errBufferLengthExceeds        = "buffer length exceeds %d"
	errCharNotExpected            = "expected '%c', got '%c'"
	errURLUnsupportedScheme       = "unsupported scheme '%s'"
	errURLOpaquedNotSupported     = "URLs with opaque data are not supported"
	errURLFragmentedNotSupported  = "URLs with fragments are not supported"
	errProtocolVersionNotExpected = "expected '%s', got %v"
	errMaxContentLengthExceeded   = "Content-Length exceeds %d (it's %d)"
	errInvalidContentLength       = "invalid Content-Length"
	errEmptyStatusMessage         = "empty status message"
	errEmptyMethod                = "empty method"
)

func ErrorStreamNotFound() error {
	return errors.New(errStreamNotFound)
}

func ErrorBadOption(o string) error {
	return fmt.Errorf(errBadOption, o)
}

func ErrorParsingOptionFail(s string) error {
	return fmt.Errorf(errParsingOption, s)
}

func ErrorParsingStreamFail(s string) error {
	return fmt.Errorf(errParsingStream, s)
}

func ErrorParsingCSeq() error {
	return errors.New(errParsingCSeq)
}

func ErrorParsingStatusCode() error {
	return errors.New(errParsingStatusCode)
}

func ErrorProtocolVersionNotExpected(s string, b []byte) error {
	return fmt.Errorf(errProtocolVersionNotExpected, s, b)
}

func ErrorCharNotExpected(cmp, byt byte) error {
	return fmt.Errorf(errCharNotExpected, cmp, byt)
}

func ErrorBufferLengthExceeds(n int) error {
	return fmt.Errorf(errBufferLengthExceeds, n)
}

func ErrorURLUnsupportedScheme(s string) error {
	return fmt.Errorf(errURLUnsupportedScheme, s)
}

func ErrorURLOpaquedNotSupported() error {
	return errors.New(errURLOpaquedNotSupported)
}

func ErrorURLFragmentedNotSupported() error {
	return errors.New(errURLFragmentedNotSupported)
}

func ErrorMaxContentLengthExceeded(l int, cl int64) error {
	return fmt.Errorf(errMaxContentLengthExceeded, l, cl)
}

func ErrorInvalidContentLength() error {
	return errors.New(errInvalidContentLength)
}

func ErrorEmptyStatusMessage() error {
	return fmt.Errorf(errEmptyStatusMessage)
}

func ErrorEmptyMethod() error {
	return errors.New(errEmptyMethod)
}
