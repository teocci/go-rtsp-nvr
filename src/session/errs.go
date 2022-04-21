// Package session
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-28
package session

import (
	"errors"
)

const (
	errStreamExitNoVideoOnStream = "stream exit, no video on stream"
	errStreamExitRtspDisconnect  = "stream exit, RTSP disconnected"
	errStreamExitNoViewer        = "stream exit, no viewer on demand"

	errStreamNotFound        = "stream not found"
	errStreamAlreadyExists   = "stream already exists"
	errChannelNotFound       = "channel not found"
	errChannelAlreadyExists  = "channel already exists"
	errChannelCodecNotFound  = "channel codec not ready, possible stream offline"
	errNoVideoOnStream       = "stream has no video"
	errNoSubscribersOnStream = "stream has no subscribers"
	errStreamRestarted       = "stream restarted"
	errStreamStopCoreSignal  = "core signal stream stopped"
	errStreamStopRTSPSignal  = "rtsp-server signal stream stopped"

	errNotHLSSegments = "hls: not ts seq found"
	errEmptyPayload   = "payload len zero"
	errEmptySDP       = "sdp len zero"

	errRemoteAuthorizationFailed = "remote authorization failed"
)

func ErrorStreamNotFound() error {
	return errors.New(errStreamNotFound)
}

func ErrorStreamAlreadyExists() error {
	return errors.New(errStreamAlreadyExists)
}

func ErrorChannelNotFound() error {
	return errors.New(errChannelNotFound)
}

func ErrorChannelAlreadyExists() error {
	return errors.New(errChannelAlreadyExists)
}

func ErrorStreamChannelCodecNotFound() error {
	return errors.New(errChannelCodecNotFound)
}

func ErrorNotHLSSegments() error {
	return errors.New(errNotHLSSegments)
}

func ErrorNoVideoOnStream() error {
	return errors.New(errNoVideoOnStream)
}

func ErrorNoSubscribersOnStream() error {
	return errors.New(errNoSubscribersOnStream)
}

func ErrorStreamRestarted() error {
	return errors.New(errStreamRestarted)
}

func ErrorStreamStopCoreSignal() error {
	return errors.New(errStreamStopCoreSignal)
}

func ErrorStreamStopRTSPSignal() error {
	return errors.New(errStreamStopRTSPSignal)
}

func ErrorEmptySDP() error {
	return errors.New(errEmptySDP)
}

func ErrorEmptyPayload() error {
	return errors.New(errEmptyPayload)
}

func ErrorStreamExitNoVideoOnStream() error {
	return errors.New(errStreamExitNoVideoOnStream)
}
func ErrorStreamExitRtspDisconnected() error {
	return errors.New(errStreamExitRtspDisconnect)
}
func ErrorStreamExitNoViewer() error {
	return errors.New(errStreamExitNoViewer)
}

func ErrorRemoteAuthorizationFailed() error {
	return errors.New(errRemoteAuthorizationFailed)
}
