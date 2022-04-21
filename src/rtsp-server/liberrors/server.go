// Package liberrors
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-08
package liberrors

import (
	"errors"
	"fmt"
	"net"
)

const (
	errTLSCantBeUsedWithUDP            = "TLS can't be used with UDP"
	errTLSCantBeUsedWithUDPMulticast   = "TLS can't be used with UDP-multicast"
	errRTSPAddressNotProvided          = "RTSPAddress not provided"
	errUDPAddressesMustBeUsedTogether  = "UDPRTPAddress and UDPRTCPAddress must be used together"
	errRTPPortMustBeEven               = "RTP port must be even"
	errRTPPortsMustBeConsecutive       = "RTP and RTCP ports must be consecutive"
	errMulticastInfoMustBeUsedTogether = "MulticastIPRange, MulticastRTPPort and MulticastRTCPPort must be used together"
)

func ErrorTLSCantBeUsedWithUDP() error {
	return errors.New(errTLSCantBeUsedWithUDP)
}

func ErrorTLSCantBeUsedWithUDPMulticast() error {
	return errors.New(errTLSCantBeUsedWithUDPMulticast)
}

func ErrorRTSPAddressNotProvided() error {
	return errors.New(errRTSPAddressNotProvided)
}

func ErrorUDPAddressesMustBeUsedTogether() error {
	return errors.New(errUDPAddressesMustBeUsedTogether)
}

func ErrorRTPPortMustBeEven() error {
	return errors.New(errRTPPortMustBeEven)
}

func ErrorRTPPortsMustBeConsecutive() error {
	return errors.New(errRTPPortsMustBeConsecutive)
}

func ErrorMulticastInfoMustBeUsedTogether() error {
	return errors.New(errMulticastInfoMustBeUsedTogether)
}

const (
	errTerminated                              = "terminated"
	errSessionNotFound                         = "session not found"
	errNoUDPPacketsInAWhile                    = "no UDP packets received in a while"
	errNoRTSPRequestsInAWhile                  = "no RTSP requests received in a while"
	errCSeqMissing                             = "CSeq is missing"
	errUnhandledRequest                        = "unhandled request: %v %v"
	errInvalidState                            = "must be in state %v, while is in state %v"
	errInvalidPath                             = "invalid path"
	errContentTypeMissing                      = "Content-Type header is missing"
	errContentTypeUnsupported                  = "unsupported Content-Type header '%v'"
	errSDPInvalid                              = "invalid SDP: %v"
	errTransportHeaderInvalid                  = "invalid transport header: %v"
	errTrackAlreadySetup                       = "track %d has already been setup"
	errTransportHeaderInvalidMode              = "transport header contains a invalid mode (%v)"
	errTransportHeaderNoClientPorts            = "transport header does not contain client ports"
	errTransportHeaderNoInterleavedIDs         = "transport header does not contain interleaved IDs"
	errTransportHeaderInvalidInterleavedIDs    = "invalid interleaved IDs"
	errTransportHeaderInterleavedIDAlreadyUsed = "interleaved IDs already used"
	errTracksDifferentProtocols                = "can't setup tracks with different protocols"
	errNoTracksSetup                           = "no tracks have been setup"
	errNotAllAnnouncedTracksSetup              = "not all announced tracks have been setup"
	errCxnLinkedToOtherSession                 = "connection is linked to another session"
	errSessionTeardown                         = "teared down by %v"
	errSessionLinkedToOtherCxn                 = "session is linked to another connection"
	errInvalidSession                          = "invalid session"
	errServerPathHasChanged                    = "path has changed, was '%s', now is '%s'"
	errCannotUseSessionCreatedByOtherIP        = "cannot use a session created with a different IP"
	errServerUDPPortsAlreadyInUse              = "UDP ports %d and %d are already in use by another reader"
	errSessionNotInUse                         = "not in use"
)

func ErrorServerTerminated() error {
	return errors.New(errTerminated)
}

func ErrorSessionNotFound() error {
	return errors.New(errSessionNotFound)
}

func ErrorNoUDPPacketsInAWhile() error {
	return errors.New(errNoUDPPacketsInAWhile)
}

func ErrorNoRTSPRequestsInAWhile() error {
	return errors.New(errNoRTSPRequestsInAWhile)
}

type ErrCSeqMissing struct{}

// Error implements the error interface.
func (e ErrCSeqMissing) Error() string {
	return errCSeqMissing
}

func ErrorCSeqMissing() error {
	return errors.New(errCSeqMissing)
}

func ErrorUnhandledRequest(m, u string) error {
	return fmt.Errorf(errUnhandledRequest, m, u)
}

func ErrorInvalidState(l []fmt.Stringer, s fmt.Stringer) error {
	return fmt.Errorf(errInvalidState, l, s)
}

func ErrorInvalidPath() error {
	return errors.New(errInvalidPath)
}

func ErrorContentTypeMissing() error {
	return errors.New(errContentTypeMissing)
}

func ErrorContentTypeUnsupported(v string) error {
	return fmt.Errorf(errContentTypeUnsupported, v)
}

func ErrorSDPInvalid(e error) error {
	return fmt.Errorf(errSDPInvalid, e)
}

func ErrorTransportHeaderInvalid(e error) error {
	return fmt.Errorf(errTransportHeaderInvalid, e)
}

func ErrorTrackAlreadySetup(ID int) error {
	return fmt.Errorf(errTrackAlreadySetup, ID)
}

func ErrorTransportHeaderInvalidMode(m int) error {
	return fmt.Errorf(errTransportHeaderInvalidMode, m)
}

func ErrorTransportHeaderNoClientPorts() error {
	return errors.New(errTransportHeaderNoClientPorts)
}

func ErrorTransportHeaderNoInterleavedIDs() error {
	return errors.New(errTransportHeaderNoInterleavedIDs)
}

func ErrorTransportHeaderInvalidInterleavedIDs() error {
	return errors.New(errTransportHeaderInvalidInterleavedIDs)
}

func ErrorTransportHeaderInterleavedIDsAlreadyUsed() error {
	return errors.New(errTransportHeaderInterleavedIDAlreadyUsed)
}

func ErrorTracksDifferentProtocols() error {
	return errors.New(errTracksDifferentProtocols)
}

func ErrorNoTracksSetup() error {
	return errors.New(errNoTracksSetup)
}

func ErrorNotAllAnnouncedTracksSetup() error {
	return errors.New(errNotAllAnnouncedTracksSetup)
}

func ErrorCxnLinkedToOtherSession() error {
	return errors.New(errCxnLinkedToOtherSession)
}

func ErrorSessionTeardown(a net.Addr) error {
	return fmt.Errorf(errSessionTeardown, a)
}

func ErrorSessionLinkedToOtherConn() error {
	return errors.New(errSessionLinkedToOtherCxn)
}

func ErrorInvalidSession() error {
	return errors.New(errInvalidSession)
}

func ErrorServerPathHasChanged(p, c string) error {
	return fmt.Errorf(errServerPathHasChanged, p, c)
}

func ErrorCannotUseSessionCreatedByOtherIP() error {
	return errors.New(errCannotUseSessionCreatedByOtherIP)
}

func ErrorServerUDPPortsAlreadyInUse(p int) error {
	return fmt.Errorf(errServerUDPPortsAlreadyInUse, p, p+1)
}

func ErrorSessionNotInUse() error {
	return errors.New(errSessionNotInUse)
}

const (
	errSetupRequestPathMustEndWithSlash = "path of a SETUP request must end with a slash. " +
		"This typically happens when VLC fails a request, and then switches to an " +
		"unsupported RTSP dialect"
	errUnableToParseTrackID                = "unable to parse track ID (%v)"
	errCannotSetupTracksWithDifferentPaths = "can't setup tracks with different paths"
	errInvalidTrackPath                    = "invalid track path (%s)"
	errUnableToGenerateTrackURL            = "unable to generate track URL"
	errInvalidTrackURL                     = "invalid track URL (%v)"
	errInvalidTrackPathMustBeginWith       = "invalid track path: must begin with '%s', but is '%s'"
)

func ErrorSetupRequestPathMustEndWithSlash() error {
	return errors.New(errSetupRequestPathMustEndWithSlash)
}

func ErrorUnableToParseTrackID(s string) error {
	return fmt.Errorf(errUnableToParseTrackID, s)
}

func ErrorCannotSetupTracksWithDifferentPaths() error {
	return errors.New(errCannotSetupTracksWithDifferentPaths)
}

func ErrorInvalidTrackPath(s string) error {
	return fmt.Errorf(errInvalidTrackPath, s)
}

func ErrorUnableToGenerateTrackURL() error {
	return errors.New(errUnableToGenerateTrackURL)
}

func ErrorInvalidTrackURL(u string) error {
	return fmt.Errorf(errInvalidTrackURL, u)
}

func ErrorInvalidTrackPathMustBeginWith(p, tp string) error {
	return fmt.Errorf(errInvalidTrackPathMustBeginWith, p, tp)
}
