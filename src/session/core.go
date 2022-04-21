package session

import (
	"github.com/teocci/go-stream-av/format/rtmp"
	"log"
	"math"
	"time"

	"github.com/teocci/go-stream-av/av"
	"github.com/teocci/go-stream-av/format/rtspv2"
)

//StreamServerRunStreamDo stream run do mux
func StreamServerRunStreamDo(streamID, channelID string) {
	var status int
	defer func() {
		//TODO fix it no need unlock run if delete stream
		if status != 2 {
			CoreSession.UnlockChannel(streamID, channelID)
		}
	}()

	log.Printf("StreamServerRunStreamDo [stream]:[%s] | [channel]:[%s]", streamID, channelID)
	log.Printf("Run stream")

	for {
		channel, err := CoreSession.GetChannel(streamID, channelID)
		if err != nil {
			log.Printf("GetChannel | Error: %s", err)
			return
		}

		if channel.OnDemand && !CoreSession.HasSubscriber(streamID, channelID) {
			log.Printf("ClientHas | Error: %s", "Stop stream no client")
			return
		}

		status, err = StreamServerRunStream(streamID, channelID, channel)
		if err != nil {
			log.Printf("Restart | Error: %s - %s", "Stream error restart stream", err)
		}
		if status > 0 {
			log.Printf("StreamServerRunStream | Error: %s", "Stream exit by signal or not client")
			return
		}

		time.Sleep(2 * time.Second)
	}
}

//StreamServerRunStream core stream
func StreamServerRunStream(streamID, channelID string, channel *Channel) (status int, err error) {
	keyTest := time.NewTimer(20 * time.Second)
	checkSubscribers := time.NewTimer(20 * time.Second)

	var started bool
	var fps int
	var preKeyTS = time.Duration(0)
	var packets []*av.Packet

	RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:                channel.URL,
		InsecureSkipVerify: channel.InsecureSkipVerify,
		DisableAudio:       !channel.Audio,
		DialTimeout:        3 * time.Second,
		ReadWriteTimeout:   5 * time.Second,
		Debug:              channel.Debug,
		OutgoingProxy:      true,
	})
	if err != nil {
		return
	}

	CoreSession.UpdateChannelStatus(streamID, channelID, ONLINE)
	defer func() {
		RTSPClient.Close()
		CoreSession.UpdateChannelStatus(streamID, channelID, OFFLINE)
		CoreSession.StreamHLSFlush(streamID, channelID)
	}()

	var waitingCodec bool
	if RTSPClient.WaitCodec {
		waitingCodec = true
	} else {
		if len(RTSPClient.CodecData) > 0 {
			CoreSession.UpdateCodecInfo(streamID, channelID, RTSPClient.CodecData, RTSPClient.SDPRaw)
		}
	}

	log.Printf("StreamServerRunStream [stream]:[%s] | [channel]:[%s] | Start: %s", streamID, channelID, "Success connection RTSP")
	var probeCount int
	var probeFrame int
	var probePTS time.Duration
	CoreSession.NewHLSMuxer(streamID, channelID)
	defer CoreSession.HLSMuxerClose(streamID, channelID)

	for {
		select {
		//Check stream have clients
		case <-checkSubscribers.C:
			if channel.OnDemand && !CoreSession.HasSubscriber(streamID, channelID) {
				return 1, ErrorNoSubscribersOnStream()
			}
			checkSubscribers.Reset(20 * time.Second)
		//Check stream send key
		case <-keyTest.C:
			return 0, ErrorNoVideoOnStream()
		//Read core Signals
		case signals := <-channel.Signals:
			switch signals {
			case SignalStreamStopped:
				return 2, ErrorStreamStopCoreSignal()
			case SignalStreamRestarted:
				return 0, ErrorStreamRestarted()
			case SignalStreamHasNoSubscribers:
				return 1, ErrorNoSubscribersOnStream()
			}
		//Read rtsp-server Signals
		case signals := <-RTSPClient.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				CoreSession.UpdateCodecInfo(streamID, channelID, RTSPClient.CodecData, RTSPClient.SDPRaw)
				waitingCodec = false
			case rtspv2.SignalStreamRTPStop:
				return 0, ErrorStreamStopRTSPSignal()
			}

		case packetRTP := <-RTSPClient.OutgoingProxyQueue:
			CoreSession.ChannelCastProxy(streamID, channelID, packetRTP)

		case packetAV := <-RTSPClient.OutgoingPacketQueue:
			if waitingCodec {
				continue
			}

			if packetAV.IsKeyFrame {
				keyTest.Reset(20 * time.Second)
				if preKeyTS > 0 {
					CoreSession.StreamHLSAdd(streamID, channelID, packets, packetAV.Time-preKeyTS)
					packets = []*av.Packet{}
				}
				preKeyTS = packetAV.Time
			}

			packets = append(packets, packetAV)

			CoreSession.ChannelCast(streamID, channelID, packetAV)

			// Low-latency HLS check
			if packetAV.IsKeyFrame && !started {
				started = true
			}

			// FPS Mode probe
			if started {
				probePTS += packetAV.Duration
				probeFrame++
				if packetAV.IsKeyFrame && probePTS.Seconds() >= 1 {
					probeCount++
					if probeCount == 2 {
						fps = int(math.Round(float64(probeFrame) / probePTS.Seconds()))
					}
					probeFrame = 0
					probePTS = 0
				}
			}

			if started && fps != 0 {
				//TODO fix it
				packetAV.Duration = time.Duration((1e3/float32(fps))*1e6) * time.Nanosecond
				CoreSession.HLSMuxerSetFPS(streamID, channelID, fps)
				CoreSession.HLSMuxerWritePacket(streamID, channelID, packetAV)
			}
		}
	}
}

func StreamServerRunStreamRTMP(streamID, channelID string, channel *Channel) (int, error) {
	keyTest := time.NewTimer(20 * time.Second)
	checkSubscribers := time.NewTimer(20 * time.Second)
	outGoingPacketQueue := make(chan *av.Packet, 1000)
	signals := make(chan int, 100)

	var started bool
	var fps int
	var preKeyTS = time.Duration(0)
	var packets []*av.Packet

	conn, err := rtmp.DialTimeout(channel.URL, 3*time.Second)
	if err != nil {
		return 0, err
	}

	CoreSession.UpdateChannelStatus(streamID, channelID, ONLINE)
	defer func() {
		conn.Close()
		CoreSession.UpdateChannelStatus(streamID, channelID, OFFLINE)
		CoreSession.StreamHLSFlush(streamID, channelID)
	}()

	var waitingCodec bool

	codecs, err := conn.Streams()
	if err != nil {
		return 0, err
	}

	durations := make([]time.Duration, len(codecs))
	CoreSession.UpdateCodecInfo(streamID, channelID, codecs, []byte{})

	log.Printf("StreamServerRunStreamRTMP [stream]:[%s] | [channel]:[%s] | Start: %s", streamID, channelID, "Success connection RTMP")
	var probeCount int
	var probeFrame int
	var probePTS time.Duration
	CoreSession.NewHLSMuxer(streamID, channelID)
	defer CoreSession.HLSMuxerClose(streamID, channelID)

	go func() {
		for {
			packet, err := conn.ReadPacket()
			if err != nil {
				break
			}
			outGoingPacketQueue <- &packet
		}
		signals <- 1
	}()

	for {
		select {
		// Check stream have clients
		case <-checkSubscribers.C:
			if channel.OnDemand && !CoreSession.HasSubscriber(streamID, channelID) {
				return 1, ErrorNoSubscribersOnStream()
			}
			checkSubscribers.Reset(20 * time.Second)
		// Check stream send key
		case <-keyTest.C:
			return 0, ErrorNoVideoOnStream()
		// Read core Signals
		case s := <-channel.Signals:
			switch s {
			case SignalStreamStopped:
				return 2, ErrorStreamStopCoreSignal()
			case SignalStreamRestarted:
				return 0, ErrorStreamRestarted()
			case SignalStreamHasNoSubscribers:
				return 1, ErrorNoSubscribersOnStream()
			}
		// Read RTSP Signals
		case <-signals:
			return 0, ErrorStreamStopRTSPSignal()
		case packetAV := <-outGoingPacketQueue:
			if durations[packetAV.Idx] != 0 {
				packetAV.Duration = packetAV.Time - durations[packetAV.Idx]
			}

			durations[packetAV.Idx] = packetAV.Time

			if waitingCodec {
				continue
			}

			if packetAV.IsKeyFrame {
				keyTest.Reset(20 * time.Second)
				if preKeyTS > 0 {
					CoreSession.StreamHLSAdd(streamID, channelID, packets, packetAV.Time-preKeyTS)
					packets = []*av.Packet{}
				}
				preKeyTS = packetAV.Time
			}
			packets = append(packets, packetAV)
			CoreSession.ChannelCast(streamID, channelID, packetAV)

			// Low-latency HLS check
			if packetAV.IsKeyFrame && !started {
				started = true
			}

			// FPS Mode probe
			if started {
				probePTS += packetAV.Duration
				probeFrame++
				if packetAV.IsKeyFrame && probePTS.Seconds() >= 1 {
					probeCount++
					if probeCount == 2 {
						fps = int(math.Round(float64(probeFrame) / probePTS.Seconds()))
					}
					probeFrame = 0
					probePTS = 0
				}
			}

			if started && fps != 0 {
				//TODO fix it
				packetAV.Duration = time.Duration((1e3/float32(fps))*1e6) * time.Nanosecond
				CoreSession.HLSMuxerSetFPS(streamID, channelID, fps)
				CoreSession.HLSMuxerWritePacket(streamID, channelID, packetAV)
			}
		}
	}
}
