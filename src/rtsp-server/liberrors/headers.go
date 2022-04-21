// Package liberrors
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-07
package liberrors

import (
	"errors"
	"fmt"
)

const (
	errValueNotFound              = "value not found (%v)"
	errValueNotProvided           = "value not provided"
	errValueProvidedMultipleTimes = "value provided multiple times (%v)"
	errInvalidValue               = "invalid value"
	errInvalidValueWithV          = "invalid value (%v)"
	errUnableToSplitMethodKey     = "unable to split between method and keys (%v)"
	errInvalidMethod              = "invalid method (%s)"
	errInvalidAuthorizationHeader = "invalid authorization header"
	errApexesNotClosed            = "apexes not closed (%v)"
	errInvalidSMPTETime           = "invalid SMPTE time (%v)"
	errInvalidNTPTime             = "invalid NPT time (%v)"

	errInvalidPorts         = "invalid ports (%v)"
	errInvalidSource        = "invalid source (%v)"
	errInvalidDestination   = "invalid destination (%v)"
	errInvalidSSRC          = "invalid SSRC"
	errInvalidTransportMode = "invalid transport mode: '%s'"
	errProtocolNotFound     = "protocol not found (%v)"
)

func ErrorValueNotFound(v string) error {
	return fmt.Errorf(errValueNotFound, v)
}

func ErrorValueNotProvided() error {
	return errors.New(errValueNotProvided)
}

func ErrorValueProvidedMultipleTimes(v []string) error {
	return fmt.Errorf(errValueProvidedMultipleTimes, v)
}

func ErrorInvalidValue() error {
	return errors.New(errInvalidValue)
}

func ErrorInvalidValueWithV(v string) error {
	return fmt.Errorf(errInvalidValueWithV, v)
}

func ErrorUnableToSplitMethodKey(s string) error {
	return fmt.Errorf(errUnableToSplitMethodKey, s)
}

func ErrorInvalidMethod(s string) error {
	return fmt.Errorf(errInvalidMethod, s)
}

func ErrorInvalidAuthorizationHeader() error {
	return fmt.Errorf(errInvalidAuthorizationHeader)
}

func ErrorApexesNotClosed(o string) error {
	return fmt.Errorf(errApexesNotClosed, o)
}

func ErrorInvalidSMPTETime(s string) error {
	return fmt.Errorf(errInvalidSMPTETime, s)
}

func ErrorInvalidNTPTime(s string) error {
	return fmt.Errorf(errInvalidNTPTime, s)
}

func ErrorInvalidPorts(val string) error {
	return fmt.Errorf(errInvalidPorts, val)
}

func ErrorInvalidSource(v string) error {
	return fmt.Errorf(errInvalidSource, v)
}

func ErrorInvalidDestination(v string) error {
	return fmt.Errorf(errInvalidDestination, v)
}

func ErrorInvalidSSRC() error {
	return errors.New(errInvalidSSRC)
}

func ErrorInvalidTransportMode(s string) error {
	return fmt.Errorf(errInvalidTransportMode, s)
}

func ErrorProtocolNotFound(s string) error {
	return fmt.Errorf(errProtocolNotFound, s)
}
