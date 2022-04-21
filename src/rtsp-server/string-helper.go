// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-14
package rtsp_server

func stringsReverseIndex(s, substr string) int {
	for i := len(s) - 1 - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}
