// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-20
package webserver

import (
	"errors"
	"fmt"
)

const (
	errEmptyPayloadRequest       = "empty payload"
	errParsingJSONRequest        = "error parsing json: %s"
	errEmptyCodec                = "empty codec"
	errRemoteAuthorizationFailed = "remote authorization failed"
)

func ErrorEmptyPayloadRequest() error {
	return errors.New(errEmptyPayloadRequest)
}

func ErrorParsingJSONRequest(e error) error {
	return fmt.Errorf(errParsingJSONRequest, e)
}

func ErrorEmptyCodec() error {
	return errors.New(errEmptyCodec)
}

func ErrorRemoteAuthorizationFailed() error {
	return errors.New(errRemoteAuthorizationFailed)
}
