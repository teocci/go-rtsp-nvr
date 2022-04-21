// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2021-Nov-01
package webserver

import (
	"github.com/gin-gonic/gin"
)

type HTMLCode int

const (
	HTMLCodeUnknown HTMLCode = 0
	HTMLCodeOK      HTMLCode = 200

	HTMLCodeBadRequest       HTMLCode = 400
	HTMLCodeUnauthorized     HTMLCode = 401
	HTMLCodeNotFound         HTMLCode = 404
	HTMLCodeMethodNotAllowed HTMLCode = 405
	HTMLCodeRequestTimeout   HTMLCode = 408
	HTMLCodeServerError      HTMLCode = 500
)

const (
	StatusCodeFail    = 0
	StatusCodeSuccess = 1

	Success = "success"
)

type ResponseMessage struct {
	Code    HTMLCode    `json:"-"`
	Status  int         `json:"status"`
	Payload interface{} `json:"payload,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ExecStatus struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type ResponseError struct {
	Error string `json:"error"`
}

type Response struct {
	Status     int
	StreamData StreamData
	Message    []string
	Error      []string
}

type WebCodec struct {
	Type string `json:"type"`
}

type StreamData struct {
	StreamID string     `json:"stream_id"`
	Tracks   []WebCodec `json:"tracks,omitempty"`
	Sdp64    string     `json:"sdp64,omitempty"`
}

func (sd StreamData) IsNil() bool {
	return len(sd.StreamID) == 0 && len(sd.Sdp64) == 0 && len(sd.Tracks) == 0
}

func SendResponse(c *gin.Context, response ResponseMessage) {
	c.JSON(int(response.Code), response)
}
