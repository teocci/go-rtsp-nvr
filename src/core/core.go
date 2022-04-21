// Package core
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-18
package core

import (
	"context"
	"fmt"
	rtspserver "github.com/teocci/go-rtsp-nvr/src/rtsp-server"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/teocci/go-rtsp-nvr/src/logger"
	"github.com/teocci/go-rtsp-nvr/src/session"
	"github.com/teocci/go-rtsp-nvr/src/webserver"
)

type InitData struct {
	Start     bool
	DroneID   int
	CompanyID int
}

// Core is an instance of rtsp-server-simple-server.
type Core struct {
	ctx       context.Context
	ctxCancel func()
	confPath  string
	confFound bool
	logger    *logger.Logger

	// out
	done chan struct{}
}

const (
	prefix           = "jinan-smati"
	defaultDroneName = "drone-04"

	formatStreamName = "%s-%s"

	streamRTSPURL = "rtsp-server://106.244.179.242:554/jinan_test"
	//streamRTSPURL      = "rtsp-server://223.171.70.110:554/video0"
	streamOnDemand     = false
	streamDisableAudio = false
	streamDebug        = false
)

var (
	// Session global
	Session = session.LoadData()
)

func Start(data InitData) error {
	if data.Start {
		// fmt.Printf("%#v\n", data)
		pid := os.Getpid()
		fmt.Println("PID:", pid)

		//stream := session.Stream{
		//	Name:           streamName,
		//	URL:            streamRTSPURL,
		//	OnDemand:       streamOnDemand,
		//	DisableAudio:   streamDisableAudio,
		//	Debug:          streamDebug,
		//	Channels: make(map[string]session.Subscriber),
		//}
		//Session.Streams[streamName] = stream

		go webserver.Start(Session)
		go rtspserver.Start(Session)
		//go streamer.ServeStreams(Session)
		go session.CoreSession.RunAllChannels()
		//go streamer.ServeSingleStream(Session, streamName)

		sigs := make(chan os.Signal, 1)
		done := make(chan bool, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			log.Println(sig)
			done <- true
		}()
		session.CoreSession.StopAllChannels()
		time.Sleep(2 * time.Second)
		log.Println("Server start awaiting signal")
		<-done
		log.Println("Server stop working by signal")
	}

	return nil
}
