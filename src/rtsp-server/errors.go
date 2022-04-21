// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-11
package rtsp_server

import (
	"errors"
	"fmt"
)

const (
	errFMTPAttributeIsMissing  = "fmtp attribute is missing"
	errParameterSetsAreMissing = "sprop-parameter-sets are missing (%v)"
	errInvalidParameterSets    = "invalid sprop-parameter-sets (%v)"
	errInvalidFMTPAttribute    = "invalid fmtp attribute (%v)"
)

func ErrorFMTPAttributeIsMissing() error {
	return errors.New(errFMTPAttributeIsMissing)
}

func ErrorParameterSetsAreMissing(v string) error {
	return fmt.Errorf(errParameterSetsAreMissing, v)
}

func ErrorInvalidParameterSets(v string) error {
	return fmt.Errorf(errInvalidParameterSets, v)
}

func ErrorInvalidFMTPAttribute(v string) error {
	return fmt.Errorf(errInvalidFMTPAttribute, v)
}
