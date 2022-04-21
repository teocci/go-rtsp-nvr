// Package session
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-05
package session

type SubscriberMode int

// Default stream type
const (
	MSE SubscriberMode = iota
	WEBRTC
	RTSP
)

// Default stream status type
const (
	OFFLINE = iota
	ONLINE
)

var (
	Success = "success"
)
