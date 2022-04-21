// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-19
package webserver

import "github.com/teocci/go-rtsp-nvr/src/session"

type ServeResponse struct {
	Version   string
	Port      string
	Template  string
	Page      string
	StreamID  string
	ChannelID string
	Streams   map[string]session.Stream
}
