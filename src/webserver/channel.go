// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-20
package webserver

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/teocci/go-rtsp-nvr/src/session"
)

// HandleChannelCodec function return codec info struct
func HandleChannelCodec(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	log.Printf("HandleChannelCodec | [stream][%s], [channel][%s]", streamID, channelID)

	if !Session.ChannelExist(streamID, channelID) {
		err := session.ErrorChannelNotFound()
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Println(err)
		return
	}

	codecs, err := Session.ChannelCodecs(streamID, channelID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ChannelCodecs | [stream][%s], [channel][%s]", streamID, channelID)
		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Payload: codecs,
	})
}

// HandleChannelInfo function return stream info struct
func HandleChannelInfo(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	info, err := Session.ChannelInfo(streamID, channelID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ChannelInfo [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Payload: info,
	})
}

// HandleReloadChannel function reload stream
func HandleReloadChannel(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	err := Session.ReloadChannel(streamID, channelID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ReloadChannel [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Message: Success,
	})
}

// HandleEditChannel function edit stream
func HandleEditChannel(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")

	var payload session.Channel
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

	err = Session.EditChannel(streamID, channelID, payload)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("StreamChannelEdit [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Message: Success,
	})
}

// HandleDeleteChannel function delete stream
func HandleDeleteChannel(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	log.Printf("HandleDeleteChannel [stream][%s], [channel][%s]", streamID, channelID)

	err := Session.DeleteChannel(streamID, channelID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("DeleteChannel [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Message: Success,
	})
}

// HandleAddChannel function add new stream
func HandleAddChannel(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	log.Printf("HandleAddChannel [stream][%s], [channel][%s]", streamID, channelID)

	var payload session.Channel
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

	err = Session.AddChannel(streamID, channelID, payload)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("AddChannel [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	SendResponse(c, ResponseMessage{
		Code:    HTMLCodeOK,
		Status:  StatusCodeSuccess,
		Message: Success,
	})
}
