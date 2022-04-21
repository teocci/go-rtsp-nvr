// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-28
package webserver

import (
	"github.com/gin-gonic/gin"
	"github.com/teocci/go-rtsp-nvr/src/session"
	"github.com/teocci/go-rtsp-nvr/src/streamer"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"path/filepath"
)

var Session *session.Session

func Start(s *session.Session) {
	Session = s

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(CORSMiddleware())

	//// TODO add support to add stream as a request
	//router.POST("/stream/receiver/:stream", APIHandleStreamWebRTC)
	//router.GET("/stream/codec/:stream", APIHandleStreamCodec)
	//router.POST("/stream/register/:stream", APIHandlerRegisterNewStream)
	//router.POST("/stream", APIHandleStreamWebRTC2)

	if Session.WebEnable() {
		router.GET("/", HandleIndex)
		router.GET("/pages/stream/list", HandlePageStreamList)
		router.GET("/pages/stream/add", HandlePageAddStream)
		router.GET("/pages/stream/edit/:stream", HandlePageEditStream)
		router.GET("/pages/player/hls/:stream/:channel", HandlePagePlayHLS)
		router.GET("/pages/player/mse/:stream/:channel", HandlePagePlayMSE)
		router.GET("/pages/player/webrtc/:stream/:channel", HandlePagePlayWebRTC)
		router.GET("/pages/multiview", HandlePageMultiview)
		router.Any("/pages/multiview/full", HandlePageFullScreenMultiView)
		router.GET("/pages/documentation", HandlePageDocumentation)
		router.GET("/pages/player/all/:stream/:channel", HandlePagePlayAll)
	}

	router.GET("/streams", HandleStreams)
	router.GET("/stream/delete/:stream", HandleDeleteStream)
	router.GET("/stream/reload/:stream", HandleReloadStream)
	router.GET("/stream/info/:stream", HandleStreamInfo)

	router.POST("/stream/add/:stream", HandleAddStream)
	router.POST("/stream/edit/:stream", HandleEditStream)

	router.POST("/streams/multi/control/add", HandleStreamsMultiControlAdd)
	router.POST("/streams/multi/control/delete", HandleStreamsMultiControlDelete)

	// Stream ChannelID elements
	router.POST("/stream/:stream/channel/add/:channel", HandleAddChannel)
	router.POST("/stream/:stream/channel/edit/:channel", HandleEditChannel)
	router.GET("/stream/:stream/channel/delete/:channel", HandleDeleteChannel)
	router.GET("/stream/:stream/channel/codec/:channel", HandleChannelCodec)
	router.GET("/stream/:stream/channel/reload/:channel", HandleReloadChannel)
	router.GET("/stream/:stream/channel/info/:channel", HandleChannelInfo)

	// HLS
	router.GET("/stream/:stream/channel/:channel/hls/live/index.m3u8", HandleStreamHLSM3U8)
	router.GET("/stream/:stream/channel/:channel/hls/live/segment/:seq/file.ts", HandleStreamHLSTS)

	// Low-Latency HLS
	router.GET("/stream/:stream/channel/:channel/ll-hls/live/index.m3u8", HandleStreamHLSLLM3U8)
	router.GET("/stream/:stream/channel/:channel/ll-hls/live/init.mp4", HandleStreamHLSLLInit)
	router.GET("/stream/:stream/channel/:channel/ll-hls/live/segment/:segment/:any", HandleStreamHLSLLM4Segment)
	router.GET("/stream/:stream/channel/:channel/ll-hls/live/fragment/:segment/:fragment/:any", HandleStreamLLHLSM4Fragment)

	// MSE
	router.GET("/stream/:stream/channel/:channel/mse", func(c *gin.Context) {
		handler := websocket.Handler(streamer.HandleStreamMSE)
		handler.ServeHTTP(c.Writer, c.Request)
	})
	router.POST("/stream/:stream/channel/:channel/webrtc", HandleStreamWebRTC)

	if Session.WebEnable() {
		dirPath := filepath.Join(Session.WebPath(), "/static")
		router.StaticFS("/static", http.Dir(dirPath))
	}

	//router.StaticFS("/static", http.Dir("web/static"))
	err := router.Run(Session.Server.Web.Port)
	if err != nil {
		log.Fatalln("Start HTTP Server error", err)
	}
}

////APIHandleStreamCodec stream codec
//func APIHandleStreamCodec(c *gin.Context) {
//	streamID := c.Param("stream-id")
//	channelID := c.Param("channel-id")
//	response := Response{}
//
//	if Session.ChannelExist(streamID, channelID) {
//		Session.RunChannel(streamID, channelID)
//
//		codecs, err := Session.ChannelCodecs(streamID, channelID)
//		if err != nil {
//			errMsg := "invalid codecs: track not found"
//			SendErrorResponse(c, response, errMsg)
//
//			return
//		}
//
//		response.StreamData = StreamData{
//			StreamID: streamID,
//		}
//		for _, codec := range codecs {
//			if codec.Type() != av.H264 && codec.Type() != av.PCM_ALAW && codec.Type() != av.PCM_MULAW && codec.Type() != av.OPUS {
//				log.Println("CheckH264Codecs Not Supported WebRTC ignore this track", codec.Type())
//				continue
//			}
//
//			if codec.Type().IsVideo() {
//				response.StreamData.Tracks = append(response.StreamData.Tracks, WebCodec{Type: "video"})
//			}
//			if codec.Type().IsAudio() {
//				response.StreamData.Tracks = append(response.StreamData.Tracks, WebCodec{Type: "audio"})
//			}
//		}
//
//		if len(response.StreamData.Tracks) > 0 {
//			response.Status = 200
//			SendResponse(c, response)
//
//			return
//		} else {
//			errMsg := "invalid webCodecs: codec not found"
//			SendErrorResponse(c, response, errMsg)
//
//			return
//		}
//	} else {
//		errMsg := "invalid streamID: stream not found"
//		SendErrorResponse(c, response, errMsg)
//
//		return
//	}
//}
//
////APIHandlerRegisterNewStream start a stream
//func APIHandlerRegisterNewStream(c *gin.Context) {
//	var data StreamRequest
//	extractJSONData(c.Request.Body, &data)
//
//	response := Response{}
//	if !data.IsNil() {
//		if _, ok := Session.Streams[data.StreamID]; !ok {
//			Session.Streams[data.StreamID] = session.Stream{
//				Name:         data.StreamID,
//				URL:          data.RtspURL,
//				OnDemand:     data.OnDemand,
//				DisableAudio: data.DisableAudio,
//				Debug:        data.Debug,
//				Channels:     make(map[string]session.ChannelID),
//			}
//		}
//
//		Session.RunIFNotRunning(data.StreamID)
//
//		response.Status = 200
//		response.StreamData = StreamData{
//			StreamID: data.StreamID,
//		}
//		SendResponse(c, response)
//	} else {
//		errMsg := "invalidData: null data"
//		SendErrorResponse(c, response, errMsg)
//	}
//}
//
////APIHandleStreamWebRTC stream video over WebRTC
//func APIHandleStreamWebRTC(c *gin.Context) {
//	var data StreamRequest
//	extractJSONData(c.Request.Body, &data)
//
//	response := Response{}
//
//	if !Session.StreamExits(data.StreamID) {
//		errMsg := fmt.Sprintf("invalid streamID: %s stream not found", data.StreamID)
//		SendErrorResponse(c, response, errMsg)
//
//		return
//	}
//
//	Session.RunIFNotRunning(data.StreamID)
//	tracks := Session.StreamTracks(data.StreamID)
//	if tracks == nil {
//		errMsg := fmt.Sprintf("track not found for stream: %s", data.StreamID)
//		SendErrorResponse(c, response, errMsg)
//
//		return
//	}
//
//	var AudioOnly bool
//	if len(tracks) == 1 && tracks[0].Type().IsAudio() {
//		AudioOnly = true
//	}
//
//	muxerWebRTC := webrtc.NewMuxer(webrtc.Options{
//		ICEServers:    Session.GetICEServers(),
//		ICEUsername:   Session.GetICEUsername(),
//		ICECredential: Session.GetICECredential(),
//		PortMin:       Session.GetWebRTCPortMin(),
//		PortMax:       Session.GetWebRTCPortMax(),
//	})
//	answer, err := muxerWebRTC.WriteHeader(tracks, data.SDP64)
//
//	if err != nil {
//		errMsg := fmt.Sprintf("[%s] write header: %s", data.StreamID, err.Error())
//		SendErrorResponse(c, response, errMsg)
//
//		return
//	}
//
//	response.Status = 200
//	response.StreamData = StreamData{
//		StreamID: data.StreamID,
//		Sdp64:    answer,
//	}
//	SendResponse(c, response)
//
//	go func() {
//		subscriberID, ch := Session.AddSubscriber(data.StreamID)
//
//		defer Session.RemoveSubscriber(data.StreamID, subscriberID)
//		defer muxerWebRTC.Close()
//
//		var videoStart bool
//		noVideo := time.NewTimer(10 * time.Second)
//		for {
//			select {
//			case <-noVideo.C:
//				log.Println("noVideo")
//				return
//			case pck := <-ch:
//				if pck.IsKeyFrame || AudioOnly {
//					noVideo.Reset(10 * time.Second)
//					videoStart = true
//				}
//				if !videoStart && !AudioOnly {
//					continue
//				}
//				err = muxerWebRTC.WritePacket(pck)
//				if err != nil {
//					log.Println("WritePacket", err)
//					return
//				}
//			}
//		}
//	}()
//}
//
//func APIHandleStreamWebRTC2(c *gin.Context) {
//	url := c.PostForm("url")
//	if _, ok := Session.Streams[url]; !ok {
//		Session.Streams[url] = session.Stream{
//			URL:      url,
//			OnDemand: true,
//			Channels: make(map[string]session.Subscriber),
//		}
//	}
//
//	Session.RunIFNotRunning(url)
//
//	codecs := Session.StreamTracks(url)
//	if codecs == nil {
//		log.Println("Stream CheckH264Codecs Not Found")
//		c.JSON(500, ResponseError{Error: Session.LastError.Error()})
//		return
//	}
//
//	muxerWebRTC := webrtc.NewMuxer(
//		webrtc.Options{
//			ICEServers: Session.GetICEServers(),
//			PortMin:    Session.GetWebRTCPortMin(),
//			PortMax:    Session.GetWebRTCPortMax(),
//		},
//	)
//
//	sdp64 := c.PostForm("sdp64")
//	answer, err := muxerWebRTC.WriteHeader(codecs, sdp64)
//	if err != nil {
//		log.Println("Muxer WriteHeader", err)
//		c.JSON(500, ResponseError{Error: err.Error()})
//		return
//	}
//
//	streamData := StreamData{
//		Sdp64: answer,
//	}
//
//	for _, codec := range codecs {
//		if isSupported(codec.Type()) {
//			log.Println("Codec not supported, the server will ignore this track", codec.Type())
//			continue
//		}
//		if codec.Type().IsVideo() {
//			streamData.Tracks = append(streamData.Tracks, WebCodec{Type: "video"})
//		} else {
//			streamData.Tracks = append(streamData.Tracks, WebCodec{Type: "audio"})
//		}
//	}
//
//	c.JSON(200, streamData)
//
//	AudioOnly := len(codecs) == 1 && codecs[0].Type().IsAudio()
//
//	go func() {
//		cid, ch := Session.AddSubscriber(url)
//		defer Session.RemoveSubscriber(url, cid)
//		defer muxerWebRTC.Close()
//
//		var videoStart bool
//		noVideo := time.NewTimer(10 * time.Second)
//		for {
//			select {
//			case <-noVideo.C:
//				log.Println("noVideo")
//				return
//			case pck := <-ch:
//				if pck.IsKeyFrame || AudioOnly {
//					noVideo.Reset(10 * time.Second)
//					videoStart = true
//				}
//				if !videoStart && !AudioOnly {
//					continue
//				}
//				err = muxerWebRTC.WritePacket(pck)
//				if err != nil {
//					log.Println("WritePacket", err)
//					return
//				}
//			}
//		}
//	}()
//}
