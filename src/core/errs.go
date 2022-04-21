// Package core
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-24
package core

import (
	"errors"
	"fmt"
)

const (
	errSessionIndexNotFound      = "session index not found"
	errSessionIndexNotNumerical  = "session index not numerical"
	errUnableToOpenCSVFile       = "unable to open %s file: %s"
	errUnableToCreateCSVFile     = "unable to create new file: %s -> %s"
	errStreamExitNoVideoOnStream = "stream exit, no video on stream"
	errStreamExitRtspDisconnect  = "stream exit, RTSP disconnected"
	errStreamExitNoViewer        = "stream exit, no viewer on demand"
)

func ErrorStreamExitNoVideoOnStream() error {
	return errors.New(errStreamExitNoVideoOnStream)
}
func ErrorStreamExitRtspDisconnected() error {
	return errors.New(errStreamExitRtspDisconnect)
}
func ErrorStreamExitNoViewer() error {
	return errors.New(errStreamExitNoViewer)
}

func ErrorSessionIndexNotFound() error {
	return errors.New(errSessionIndexNotFound)
}

func ErrorSessionIndexNotNumerical() error {
	return errors.New(errSessionIndexNotNumerical)
}

func ErrorUnableToOpenCSVFile(f, e string) error {
	return errors.New(fmt.Sprintf(errUnableToOpenCSVFile, f, e))
}

func ErrorUnableToCreateCSVFile(f, e string) error {
	return errors.New(fmt.Sprintf(errUnableToCreateCSVFile, f, e))
}
