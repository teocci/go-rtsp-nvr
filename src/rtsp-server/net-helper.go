// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-08
package rtsp_server

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

const (
	formatHostPort = "%s:%s"
)

func extractPort(address string) (int, error) {
	_, tmp, err := net.SplitHostPort(address)
	if err != nil {
		return 0, err
	}

	tmp2, err := strconv.ParseInt(tmp, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(tmp2), nil
}

func mergeHostPortInt(h string, p int) string {
	return mergeHostPort(h, strconv.FormatInt(int64(p), 10))
}

func mergeHostPort(h, p string) string {
	return fmt.Sprintf(formatHostPort, h, p)
}

func newSessionSecretID(sessions map[string]*ServerSession) (string, error) {
	for {
		b := make([]byte, 4)
		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}

		id := strconv.FormatUint(uint64(binary.LittleEndian.Uint32(b)), 10)

		if _, ok := sessions[id]; !ok {
			return id, nil
		}
	}
}
