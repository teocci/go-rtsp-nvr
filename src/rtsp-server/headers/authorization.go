// Package headers
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-07
package headers

import (
	"encoding/base64"
	"strings"

	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/base"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
)

// Authorization is an Authorization header.
type Authorization struct {
	// authentication method
	Method AuthMethod

	// basic user
	BasicUser string

	// basic password
	BasicPass string

	// digest values
	DigestValues Authenticate
}

// Read decodes an Authorization header.
func (h *Authorization) Read(hv base.HeaderValue) error {
	if len(hv) == 0 {
		return liberrors.ErrorValueNotProvided()
	}

	if len(hv) > 1 {
		return liberrors.ErrorValueProvidedMultipleTimes(hv)
	}

	v0 := hv[0]

	switch {
	case strings.HasPrefix(v0, "Basic "):
		h.Method = AuthBasic

		v0 = v0[len("Basic "):]

		basicPL, err := base64.StdEncoding.DecodeString(v0)
		if err != nil {
			return liberrors.ErrorInvalidValue()
		}

		tmp2 := strings.Split(string(basicPL), ":")
		if len(tmp2) != 2 {
			return liberrors.ErrorInvalidValue()
		}

		h.BasicUser, h.BasicPass = tmp2[0], tmp2[1]

	case strings.HasPrefix(v0, "Digest "):
		h.Method = AuthDigest

		var digest Authenticate
		err := digest.Read(base.HeaderValue{v0})
		if err != nil {
			return err
		}

		h.DigestValues = digest

	default:
		return liberrors.ErrorInvalidAuthorizationHeader()
	}

	return nil
}

// Write encodes an Authorization header.
func (h Authorization) Write() base.HeaderValue {
	switch h.Method {
	case AuthBasic:
		response := base64.StdEncoding.EncodeToString([]byte(h.BasicUser + ":" + h.BasicPass))

		return base.HeaderValue{"Basic " + response}

	default: // AuthDigest
		return h.DigestValues.Write()
	}
}
