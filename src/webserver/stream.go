// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-19
package webserver

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/teocci/go-rtsp-nvr/src/session"
)

// HandleStreams function return stream list
func HandleStreams(c *gin.Context) {
	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Payload: Session.StreamsList(),
	})
}

// HandleAddStream function add new stream
func HandleAddStream(c *gin.Context) {
	streamID := c.Param("stream")

	var payload session.Stream
	err := c.BindJSON(&payload)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Println(ErrorParsingJSONRequest(err))

		return
	}

	err = Session.AddStream(streamID, payload)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("AddStream [stream][%s] | [error]: %s", streamID, err)

		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Message: Success,
	})
}

// HandleEditStream function edit stream
func HandleEditStream(c *gin.Context) {
	streamID := c.Param("stream")

	var payload session.Stream
	err := c.BindJSON(&payload)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Println(ErrorParsingJSONRequest(err))

		return
	}
	err = Session.EditStream(streamID, payload)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("EditStream [stream][%s] | [error]: %s", streamID, err)

		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Message: Success,
	})
}

// HandleDeleteStream function delete stream
func HandleDeleteStream(c *gin.Context) {
	streamID := c.Param("stream")
	err := Session.DeleteStream(streamID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("DeleteStream [stream][%s] | [error]: %s", streamID, err)

		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Message: Success,
	})
}

// HandleReloadStream reload stream
func HandleReloadStream(c *gin.Context) {
	streamID := c.Param("stream")
	err := Session.StreamReload(streamID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("StreamReload [stream][%s] | [error]: %s", streamID, err)

		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Message: Success,
	})
}

// HandleStreamInfo function return stream info struct
func HandleStreamInfo(c *gin.Context) {
	streamID := c.Param("stream")
	stream, err := Session.StreamInfo(streamID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("StreamInfo [stream][%s] | [error]: %s", streamID, err)

		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Payload: stream,
	})
}

// HandleStreamsMultiControlAdd function add new stream's
func HandleStreamsMultiControlAdd(c *gin.Context) {
	streamID := c.Param("stream")
	log.Printf("HandleStreamsMultiControlAdd [stream][%s]", streamID)

	var payload session.Session
	err := c.BindJSON(&payload)
	if err != nil {
		err = ErrorParsingJSONRequest(err)
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Println(err)

		return
	}

	if payload.Streams == nil || len(payload.Streams) < 1 {
		err = ErrorEmptyPayloadRequest()
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Println(err)

		return
	}

	res := ResponseMessage{
		Code:   HTMLCodeOK,
		Status: StatusCodeSuccess,
	}
	statuses := make(map[string]ExecStatus)
	for i, stream := range payload.Streams {
		err = Session.AddStream(i, stream)
		if err != nil {
			log.Printf("AddStream [stream][%s] | [error]: %s", streamID, err)
			statuses[i] = ExecStatus{Status: 0, Message: err.Error()}
			res.Status = StatusCodeFail
		} else {
			statuses[i] = ExecStatus{Status: 1, Message: Success}
		}
	}
	res.Payload = statuses

	SendResponse(c, res)
}

// HandleStreamsMultiControlDelete function delete stream's
func HandleStreamsMultiControlDelete(c *gin.Context) {
	streamID := c.Param("stream")
	log.Printf("HandleStreamsMultiControlDelete [stream][%s]", streamID)

	var payload []string
	err := c.BindJSON(&payload)
	if err != nil {
		err = ErrorParsingJSONRequest(err)
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Println(err)

		return
	}

	if len(payload) < 1 {
		err = ErrorEmptyPayloadRequest()
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Println(err)

		return
	}

	res := ResponseMessage{
		Code:   HTMLCodeOK,
		Status: StatusCodeSuccess,
	}
	statuses := make(map[string]ExecStatus)
	for _, key := range payload {
		err = Session.DeleteStream(key)
		if err != nil {
			log.Printf("DeleteStream [stream][%s] | [error]: %s", streamID, err)
			statuses[key] = ExecStatus{Status: 0, Message: err.Error()}
			res.Status = StatusCodeFail
		} else {
			statuses[key] = ExecStatus{Status: 1, Message: Success}
		}
	}
	res.Payload = statuses

	SendResponse(c, res)
}
