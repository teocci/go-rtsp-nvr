// Package headers
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-07
package headers

import (
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
)

func readKey(s string, separator byte) (string, string) {
	i := 0
	for {
		if i >= len(s) || s[i] == '=' || s[i] == separator {
			break
		}

		i++
	}

	return s[:i], s[i:]
}

func readValue(o, s string, separator byte) (string, string, error) {
	if len(s) > 0 && s[0] == '"' {
		i := 1
		for {
			if i >= len(s) {
				return "", "", liberrors.ErrorApexesNotClosed(o)
			}

			if s[i] == '"' {
				return s[1:i], s[i+1:], nil
			}

			i++
		}
	}

	i := 0
	for {
		if i >= len(s) || s[i] == separator {
			return s[:i], s[i:], nil
		}
		i++
	}
}

func keyValParse(s string, separator byte) (ret map[string]string, err error) {
	ret = make(map[string]string)
	o := s

	for len(s) > 0 {
		var k string
		k, s = readKey(s, separator)

		if len(k) > 0 {
			if len(s) > 0 && s[0] == '=' {
				var v string
				v, s, err = readValue(o, s[1:], separator)
				if err != nil {
					return nil, err
				}

				ret[k] = v
			} else {
				ret[k] = ""
			}
		}

		// skip separator
		if len(s) > 0 && s[0] == separator {
			s = s[1:]
		}

		// skip spaces
		for len(s) > 0 && s[0] == ' ' {
			s = s[1:]
		}
	}

	return ret, nil
}
