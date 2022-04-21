// Package session
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-01
package session

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type AuthorizationRequest struct {
	Proto   string `json:"proto,omitempty"`
	Stream  string `json:"stream,omitempty"`
	Channel string `json:"channel,omitempty"`
	Token   string `json:"token,omitempty"`
	IP      string `json:"ip,omitempty"`
}

type AuthorizationResponse struct {
	Status string `json:"status,omitempty"`
}

func RemoteAuthorization(proto string, stream string, channel string, token string, ip string) bool {
	if !CoreSession.WebTokenEnable() {
		return true
	}

	buf, err := json.Marshal(&AuthorizationRequest{proto, stream, channel, token, ip})
	if err != nil {
		return false
	}

	request, err := http.NewRequest("POST", CoreSession.WebTokenBackend(), bytes.NewBuffer(buf))
	if err != nil {
		return false
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false
	}

	var authorization AuthorizationResponse
	err = json.Unmarshal(bodyBytes, &authorization)
	if err != nil {
		return false
	}

	return authorization.Status == "1"
}
