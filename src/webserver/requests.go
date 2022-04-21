// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2021-Nov-02
package webserver

type StreamRequest struct {
	StreamID     string `json:"stream_id"`
	ChannelID    string `json:"channel_id,omitempty"`
	RtspURL      string `json:"rtsp_url,omitempty"`
	OnDemand     bool   `json:"on_demand,omitempty"`
	DisableAudio bool   `json:"disable_audio,omitempty"`
	Debug        bool   `json:"debug,omitempty"`
	SDP64        string `json:"data,omitempty"`
}

func (ssr *StreamRequest) IsNil() bool {
	return len(ssr.StreamID) == 0 && len(ssr.RtspURL) == 0 && len(ssr.SDP64) == 0
}
