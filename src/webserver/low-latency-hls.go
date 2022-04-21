// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-20
package webserver

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/teocci/go-rtsp-nvr/src/session"
	"github.com/teocci/go-rtsp-nvr/src/webauth"
	"github.com/teocci/go-stream-av/format/mp4f"
)

// HandleStreamHLSLLInit send client ts segment
func HandleStreamHLSLLInit(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	token := c.Param("token")
	log.Printf("HandleStreamHLSLLInit | [stream][%s], [channel][%s]", streamID, channelID)

	if Session.WebTokenEnable() && !webauth.RemoteAuthorization(webauth.AuthRequest{
		Proto:   "HLS",
		Stream:  streamID,
		Channel: channelID,
		Token:   token,
		IP:      c.ClientIP(),
	}, Session.WebTokenBackend()) {
		err := ErrorRemoteAuthorizationFailed()
		log.Printf("RemoteAuthorization [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	if !Session.ChannelExist(streamID, channelID) {
		err := session.ErrorChannelNotFound()
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ChannelExist [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	c.Header("Content-Type", "application/x-mpegURL")

	Session.RunChannel(streamID, channelID)

	codecs, err := Session.ChannelCodecs(streamID, channelID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ChannelCodecs [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	Muxer := mp4f.NewMuxer(nil)
	err = Muxer.WriteHeader(codecs)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("Muxer.WriteHeader [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	c.Header("Content-Type", "video/mp4")
	_, buf := Muxer.GetInit(codecs)
	_, err = c.Writer.Write(buf)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ResponseWriter.Write [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}
}

// HandleStreamHLSLLM3U8 send client m3u8 play list
func HandleStreamHLSLLM3U8(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	log.Printf("HandleStreamHLSLLM3U8 | [stream][%s], [channel][%s]", streamID, channelID)

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

	c.Header("Content-Type", "application/x-mpegURL")
	Session.RunChannel(streamID, channelID)

	msm := stringToInt(c.DefaultQuery("_HLS_msn", "-1"))
	part := stringToInt(c.DefaultQuery("_HLS_part", "-1"))
	index, err := Session.HLSMuxerM3U8(streamID, channelID, msm, part)
	if err != nil {
		log.Printf("HLSMuxerM3U8 [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	_, err = c.Writer.Write([]byte(index))
	if err != nil {
		log.Printf("ResponseWriter.Write [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}
}

// HandleStreamHLSLLM4Segment send client ts segment
func HandleStreamHLSLLM4Segment(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	log.Printf("HandleStreamHLSLLM4Segment | [stream][%s], [channel][%s]", streamID, channelID)

	c.Header("Content-Type", "video/mp4")
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
		log.Printf("ChannelCodecs [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	if codecs == nil {
		err = ErrorEmptyCodec()
		log.Println(err)
		return
	}

	Muxer := mp4f.NewMuxer(nil)
	err = Muxer.WriteHeader(codecs)
	if err != nil {
		log.Printf("Muxer.WriteHeader [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	seqData, err := Session.HLSMuxerSegment(streamID, channelID, stringToInt(c.Param("segment")))
	if err != nil {
		log.Printf("Session.HLSMuxerSegment [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}
	for _, v := range seqData {
		err = Muxer.WritePacket4(*v)
		if err != nil {
			log.Printf("Muxer.WritePacket4 [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
			return
		}
	}

	buf := Muxer.Finalize()
	_, err = c.Writer.Write(buf)
	if err != nil {
		log.Printf("ResponseWriter.Write [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}
}

// HandleStreamLLHLSM4Fragment send client ts segment
func HandleStreamLLHLSM4Fragment(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	segment := stringToInt(c.Param("segment"))
	fragment := stringToInt(c.Param("fragment"))

	log.Printf("HandleStreamLLHLSM4Fragment | [stream][%s], [channel][%s]", streamID, channelID)

	c.Header("Content-Type", "video/mp4")
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
		log.Printf("ChannelCodecs [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}
	if codecs == nil {
		err = ErrorEmptyCodec()
		log.Println(err)
		return
	}

	Muxer := mp4f.NewMuxer(nil)
	err = Muxer.WriteHeader(codecs)
	if err != nil {
		log.Printf("Muxer.WriteHeader [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	seqData, err := Session.HLSMuxerFragment(streamID, channelID, segment, fragment)
	if err != nil {
		log.Printf("Session.HLSMuxerFragment [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}
	for _, v := range seqData {
		err = Muxer.WritePacket4(*v)
		if err != nil {
			log.Printf("Muxer.WritePacket4 [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
			return
		}
	}

	buf := Muxer.Finalize()
	_, err = c.Writer.Write(buf)
	if err != nil {
		log.Printf("ResponseWriter.Write [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}
}
