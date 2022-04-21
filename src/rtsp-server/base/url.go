// Package base
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-06
package base

import (
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
	"net/url"
)

// URL is a RTSP URL.
// This is basically an HTTP URL with some additional functions to handle
// control attributes.
type URL url.URL

// ParseURL parses a RTSP URL.
func ParseURL(s string) (*URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "rtsp" && u.Scheme != "rtsps" {
		return nil, liberrors.ErrorURLUnsupportedScheme(u.Scheme)
	}

	if u.Opaque != "" {
		return nil, liberrors.ErrorURLOpaquedNotSupported()
	}

	if u.Fragment != "" {
		return nil, liberrors.ErrorURLFragmentedNotSupported()
	}

	return (*URL)(u), nil
}

// String implements fmt.Stringer.
func (u *URL) String() string {
	return (*url.URL)(u).String()
}

// Clone clones a URL.
func (u *URL) Clone() *URL {
	return (*URL)(&url.URL{
		Scheme:     u.Scheme,
		User:       u.User,
		Host:       u.Host,
		Path:       u.Path,
		RawPath:    u.RawPath,
		ForceQuery: u.ForceQuery,
		RawQuery:   u.RawQuery,
	})
}

// CloneWithoutCredentials clones a URL without its credentials.
func (u *URL) CloneWithoutCredentials() *URL {
	return (*URL)(&url.URL{
		Scheme:     u.Scheme,
		Host:       u.Host,
		Path:       u.Path,
		RawPath:    u.RawPath,
		ForceQuery: u.ForceQuery,
		RawQuery:   u.RawQuery,
	})
}

// RTSPPathAndQuery returns the path and query of a RTSP URL.
func (u *URL) RTSPPathAndQuery() (pathAndQuery string, found bool) {
	if u.RawPath != "" {
		pathAndQuery = u.RawPath
	} else {
		pathAndQuery = u.Path
	}

	if u.RawQuery != "" {
		pathAndQuery += "?" + u.RawQuery
	}

	// remove leading slash
	if len(pathAndQuery) == 0 || pathAndQuery[0] != '/' {
		return
	}

	pathAndQuery = pathAndQuery[1:]

	return pathAndQuery, true
}
