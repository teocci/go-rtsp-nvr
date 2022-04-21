// Package cmdapp
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-27
package cmdapp

const (
	Name  = "nvr"
	Short = "Receives a RTSP stream and save it as mp4 and re-stream the rtsp-server as webrtc"
	Long  = `Receives a RTSP stream and save it as mp4 file. Re-streams the RTSP as a WebRTC stream`

	SName  = "start"
	SShort = "s"
	SDesc  = "Start service"

	DName  = "drone"
	DShort = "d"
	DDesc  = "Drone ID"

	CName  = "company"
	CShort = "C"
	CDesc  = "Company ID"
)

const (
	VersionTemplate = "%s %s.%s\n"
	Version         = "v1.0"
	Commit          = "0"
)
