// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-11
package rtsp_server

import (
	"time"
)

func emptyTimer() *time.Timer {
	t := time.NewTimer(0)
	<-t.C

	return t
}
