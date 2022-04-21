// Package streamer
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-28
package streamer

import (
	"errors"
)

const (
	errStreamExitNoVideoOnStream = "stream exit, no video on stream"
	errStreamExitRtspDisconnect  = "stream exit, RTSP disconnected"
	errStreamExitNoViewer        = "stream exit, no viewer on demand"
	errStreamNotFound            = "stream not found"
	errStreamHasNoVideo          = "stream has no video"
	errChannelNotFound           = "channel not found"
	errRemoteAuthorizationFailed = "remote authorization failed"
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

func ErrorStreamNotFound() error {
	return errors.New(errStreamNotFound)
}

func ErrorStreamHasNoVideo() error {
	return errors.New(errStreamHasNoVideo)
}

func ErrorChannelNotFound() error {
	return errors.New(errChannelNotFound)
}

func ErrorRemoteAuthorizationFailed() error {
	return errors.New(errRemoteAuthorizationFailed)
}
