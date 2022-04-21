// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-20
package webserver

import (
	"github.com/teocci/go-rtsp-nvr/src/streamer"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/teocci/go-rtsp-nvr/src/session"
	webrtc "github.com/teocci/go-stream-av/format/webrtcv3"
)

// HandleStreamWebRTC stream video over WebRTC
func HandleStreamWebRTC(c *gin.Context) {
	streamID := c.Param("stream")
	channelID := c.Param("channel")
	log.Printf("HandleStreamHLSM3U8 | [stream][%s], [channel][%s]", streamID, channelID)

	if !streamer.Session.ChannelExist(streamID, channelID) {
		err := session.ErrorChannelNotFound()
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ChannelExist [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	streamer.Session.RunChannel(streamID, channelID)

	codecs, err := streamer.Session.ChannelCodecs(streamID, channelID)
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeBadRequest,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ChannelCodecs [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	muxerWebRTC := webrtc.NewMuxer(webrtc.Options{
		ICEServers:    streamer.Session.ICEServers(),
		ICEUsername:   streamer.Session.ICEUsername(),
		ICECredential: streamer.Session.ICECredential(),
		PortMin:       streamer.Session.WebRTCPortMin(),
		PortMax:       streamer.Session.WebRTCPortMax(),
	})
	answer, err := muxerWebRTC.WriteHeader(codecs, c.PostForm("data"))
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("Muxer.WriteHeader [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}

	_, err = c.Writer.Write([]byte(answer))
	if err != nil {
		SendResponse(c, ResponseMessage{
			Code:   HTMLCodeServerError,
			Status: StatusCodeFail,
			Error:  err.Error(),
		})
		log.Printf("ResponseWriter.Write [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
		return
	}
	go func() {
		subscriber, err := streamer.Session.AddSubscriber(streamID, channelID, session.WEBRTC)
		if err != nil {
			SendResponse(c, ResponseMessage{
				Code:   HTMLCodeBadRequest,
				Status: StatusCodeFail,
				Error:  err.Error(),
			})
			log.Printf("AddSubscriber [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
			return
		}
		defer streamer.Session.RemoveSubscriber(streamID, channelID, subscriber.ID)

		var videoStart bool
		noVideo := time.NewTimer(10 * time.Second)
		for {
			select {
			case <-noVideo.C:
				log.Println(streamer.ErrorStreamHasNoVideo())
				return
			case pck := <-subscriber.Packet:
				if pck.IsKeyFrame {
					noVideo.Reset(10 * time.Second)
					videoStart = true
				}

				if !videoStart {
					continue
				}

				err = muxerWebRTC.WritePacket(*pck)
				if err != nil {
					log.Printf("Muxer.WritePacket [stream][%s], [channel][%s] | [error]: %s", streamID, channelID, err)
					return
				}
			}
		}
	}()
}
