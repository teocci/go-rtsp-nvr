// Package hls
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-28
package hls

import (
	"errors"
)

const (
	errStreamNotFound = "stream not found"
)

func ErrorStreamNotFound() error {
	return errors.New(errStreamNotFound)
}
