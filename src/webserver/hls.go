// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-20
package webserver

import (
	"bytes"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/teocci/go-rtsp-nvr/src/session"
	"github.com/teocci/go-rtsp-nvr/src/webauth"
	"github.com/teocci/go-stream-av/format/ts"
)

// HandleStreamHLSM3U8 send client m3u8 play list
func HandleStreamHLSM3U8(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	token := c.Param("token")
	log.Printf("HandleStreamHLSM3U8 | [stream][%s], [channel][%s]", streamID, channelID)

	if Session.WebTokenEnable() && !webauth.RemoteAuthorization(webauth.AuthRequest{
		Proto:   "HLS",
		Stream:  streamID,
		Channel: channelID,
		Token:   token,
		IP:      c.ClientIP(),
	}, Session.WebTokenBackend()) {
		err := session.ErrorStreamNotFound()
		log.Printf("RemoteAuthorization [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	if !Session.ChannelExist(streamID, channelID) {
		err := session.ErrorChannelNotFound()
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ChannelExist [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	c.Header("Content-Type", "application/x-mpegURL")
	Session.RunChannel(streamID, channelID)
	// If stream mode on_demand need wait ready segment's
	for i := 0; i < 40; i++ {
		index, seq, err := Session.StreamHLSm3u8(streamID, channelID)
		if err != nil {
			SendResponse(c, ResponseMessage{
				Code:   HTMLCodeBadRequest,
				Status: StatusCodeFail,
				Error:  err.Error(),
			})
			log.Printf("StreamHLSm3u8 [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
			return
		}

		if seq >= 6 {
			_, err = c.Writer.Write([]byte(index))
			if err != nil {
				SendResponse(c, ResponseMessage{
					Code:   HTMLCodeServerError,
					Status: StatusCodeFail,
					Error:  err.Error(),
				})
				log.Printf("ResponseWriter.Write [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
				return
			}
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// HandleStreamHLSTS send client ts segment
func HandleStreamHLSTS(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	seq := stringToInt(c.Param("seq"))

	log.Printf("HandleStreamHLSM3U8 | [stream][%s], [channel][%s]", streamID, channelID)

	if !Session.ChannelExist(streamID, channelID) {
		err := session.ErrorChannelNotFound()
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ChannelExist [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	codecs, err := Session.ChannelCodecs(streamID, channelID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ChannelCodecs [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	outfile := bytes.NewBuffer([]byte{})
	Muxer := ts.NewMuxer(outfile)
	Muxer.PaddingToMakeCounterCont = true
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

	seqData, err := Session.StreamHLSTS(streamID, channelID, seq)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("StreamHLSTS [stream][%s], [channel][%s], [seq][%d] | [error]: %s", streamID, channelID, seq, err)
		return
	}

	if len(seqData) == 0 {
		err = session.ErrorNotHLSSegments()
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("StreamHLSTS [stream][%s], [channel][%s], [seq][%d] | [error]: %s", streamID, channelID, seq, err)
		return
	}

	for _, v := range seqData {
		v.CompositionTime = 1
		err = Muxer.WritePacket(*v)
		if err != nil {
			SendResponse(c, ResponseMessage{
				Code:   HTMLCodeServerError,
				Status: StatusCodeFail,
				Error:  err.Error(),
			})
			log.Printf("Muxer.WritePacket [stream][%s], [channel][%s], [seq][%d] | [error]: %s", streamID, channelID, seq, err)
			return
		}
	}

	err = Muxer.WriteTrailer()
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("Muxer.WriteTrailer [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)

		return
	}

	_, err = c.Writer.Write(outfile.Bytes())
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
