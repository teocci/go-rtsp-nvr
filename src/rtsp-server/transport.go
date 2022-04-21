// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-07
package rtsp_server

// Transport is a RTSP transport protocol.
type Transport int

// standard transport protocols.
const (
	TransportUDP Transport = iota
	TransportUDPMulticast
	TransportTCP
)

var transportLabels = map[Transport]string{
	TransportUDP:          "UDP",
	TransportUDPMulticast: "UDP-multicast",
	TransportTCP:          "TCP",
}

// String implements fmt.Stringer.
func (t Transport) String() string {
	if l, ok := transportLabels[t]; ok {
		return l
	}

	return "unknown"
}
