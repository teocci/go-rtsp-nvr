// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-07
package rtsp_server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransportString(t *testing.T) {
	tr := TransportUDPMulticast
	require.NotEqual(t, "unknown", tr.String())

	tr = Transport(15)
	require.Equal(t, "unknown", tr.String())
}
