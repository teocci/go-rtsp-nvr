// Package streamer
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-27
package streamer

import (
	"log"
	"time"

	"github.com/teocci/go-rtsp-nvr/src/session"
	"github.com/teocci/go-stream-av/format/rtspv2"
)

var (
	Session *session.Session
)

func ServeStreams(s *session.Session) {
	Session = s
	s.LooperFn = RTSPWorkerLoop
	for streamID, stream := range Session.Streams {
		for channelID, channel := range stream.Channels {
			if !channel.OnDemand {
				channel.ID = channelID
				channel.StreamID = streamID
				go RTSPWorkerLoop(channel)
			}
		}
	}
}

func ServeSingleStream(s *session.Session, streamID, channelID string) {
	Session = s
	Session.LooperFn = RTSPWorkerLoop
	stream := Session.Streams[streamID]
	channel := stream.Channels[channelID]

	go RTSPWorkerLoop(channel)
}

func RTSPWorkerLoop(channel session.Channel) {
	defer Session.UnlockChannel(channel.StreamID, channel.ID)

	for {
		log.Println("Connecting to channel: ", channel.ID)
		err := RTSPWorker(channel)
		if err != nil {
			log.Println(err)
			Session.LastError = err
		}

		if channel.OnDemand && !Session.Stream(channel.StreamID).HasChannel(channel.ID) {
			log.Println(ErrorStreamExitNoViewer())
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func RTSPWorker(channel session.Channel) error {
	keyTest := time.NewTimer(20 * time.Second)
	clientTest := time.NewTimer(20 * time.Second)

	//add next TimeOut
	rtspClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              channel.URL,
		DisableAudio:     channel.Audio,
		DialTimeout:      3 * time.Second,
		ReadWriteTimeout: 3 * time.Second,
		Debug:            channel.Debug,
	})
	if err != nil {
		return err
	}
	defer rtspClient.Close()

	log.Println("Connected to channel: ", channel.ID)
	if rtspClient.CodecData != nil {
		Session.UpdateCodecInfo(channel.StreamID, channel.ID, rtspClient.CodecData, rtspClient.SDPRaw)
	}

	AudioOnly := rtspClient.IsAudioOnly()

	for {
		select {
		case <-clientTest.C:
			if channel.OnDemand {
				if !Session.HasSubscriber(channel.ID, channel.ID) {
					return ErrorStreamExitNoViewer()
				} else {
					clientTest.Reset(20 * time.Second)
				}
			}
		case <-keyTest.C:
			return ErrorStreamExitNoVideoOnStream()
		case signals := <-rtspClient.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				Session.UpdateCodecInfo(channel.StreamID, channel.ID, rtspClient.CodecData, rtspClient.SDPRaw)
			case rtspv2.SignalStreamRTPStop:
				return ErrorStreamExitRtspDisconnected()
			}
		case packetAV := <-rtspClient.OutgoingPacketQueue:
			if AudioOnly || packetAV.IsKeyFrame {
				keyTest.Reset(20 * time.Second)
			}
			Session.ChannelCast(channel.StreamID, channel.ID, packetAV)
		}
	}
}
