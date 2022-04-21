// Package liberrors
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-14
package liberrors

import (
	"errors"
	"fmt"
)

func ErrorClientTerminated() error {
	return errors.New(errTerminated)
}

const (
	errRTPInfoInvalid                 = "invalid RTP-Info: %v"
	errTCPTimeout                     = "TCP timeout"
	errUDPTimeout                     = "UDP timeout"
	errTransportHeaderNoDestination   = "transport header does not contain a destination"
	errTransportHeaderNoPorts         = "transport header does not contain ports"
	errTransportHeaderInvalidDelivery = "transport header contains an invalid delivery value"
	errServerPortsNotProvided         = "server ports have not been provided. Use AnyPortEnable to communicate with this server"
	errUDPPortsNotConsecutive         = "rtcpPort must be rtpPort + 1"
	errUDPPortsZero                   = "rtpPort and rtcpPort must be both zero or non-zero"
	errCannotSetupTracksDifferentURLs = "cannot setup tracks with different base URLs"
	errCannotReadPublishAtSameTime    = "cannot read and publish at the same time"
	errBadStatusCode                  = "bad status code: %d (%s)"
	SessionHeaderInvalid              = "invalid session header: %v"
)

func ErrorSessionHeaderInvalid(e error) error {
	return fmt.Errorf(SessionHeaderInvalid, e)
}

func ErrorBadStatusCode(c int, s string) error {
	return fmt.Errorf(errBadStatusCode, c, s)
}

func ErrorCannotReadPublishAtSameTime() error {
	return errors.New(errCannotReadPublishAtSameTime)
}

func ErrorCannotSetupTracksDifferentURLs() error {
	return errors.New(errCannotSetupTracksDifferentURLs)
}

func ErrorUDPPortsZero() error {
	return errors.New(errUDPPortsZero)
}

func ErrorUDPPortsNotConsecutive() error {
	return errors.New(errUDPPortsNotConsecutive)
}

func ErrorServerPortsNotProvided() error {
	return errors.New(errServerPortsNotProvided)
}

func ErrorTransportHeaderInvalidDelivery() error {
	return errors.New(errTransportHeaderInvalidDelivery)
}

func ErrorTransportHeaderNoPorts() error {
	return errors.New(errTransportHeaderNoPorts)
}

func ErrorTransportHeaderNoDestination() error {
	return errors.New(errTransportHeaderNoDestination)
}

func ErrorUDPTimeout() error {
	return errors.New(errUDPTimeout)
}

func ErrorTCPTimeout() error {
	return errors.New(errTCPTimeout)
}

func ErrorRTPInfoInvalid(e string) error {
	return fmt.Errorf(errRTPInfoInvalid, e)
}
