// Package session
// Created by RTT.
// Author: teocci@yandex.com on 2021-Oct-27
package session

import (
	"sync"
)

// Session struct
type Session struct {
	mutex     sync.RWMutex
	Server    Server            `json:"server"`
	Streams   map[string]Stream `json:"streams"`
	LooperFn  Looper
	LastError error
}

type Looper func(Channel)

// StreamsList list all stream
func (s *Session) StreamsList() (list map[string]Stream) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	list = make(map[string]Stream)
	for i, stream := range s.Streams {
		list[i] = stream
	}

	return
}

func (s *Session) Stream(streamID string) *Stream {
	if stream, ok := s.Streams[streamID]; ok {
		return &stream
	}

	return nil
}

// StopAllChannels stop stream
func (s *Session) StopAllChannels() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, stream := range s.Streams {
		stream.StopChannels()
	}
}

// StreamInfo returns a stream
func (s *Session) StreamInfo(streamID string) (stream *Stream, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stream = s.Stream(streamID)
	if stream == nil {
		err = ErrorStreamNotFound()
	}

	return
}

// StreamReload reload stream
func (s *Session) StreamReload(streamID string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if stream, ok := s.Streams[streamID]; ok {
		stream.RestartChannels()

		return nil
	}

	return ErrorStreamNotFound()
}

// DeleteStream stream
func (s *Session) DeleteStream(streamID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, ok := s.Streams[streamID]
	if !ok {
		return ErrorStreamNotFound()
	}

	for _, channel := range stream.Channels {
		channel.End()
	}
	delete(s.Streams, streamID)

	err := s.SaveConfig()
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) InitNilStreams() {
	if s.Streams == nil {
		s.Streams = make(map[string]Stream)
	}
}

////StreamInfo return stream info
//func (s *Session) StreamInfo(streamID string) (*Stream, error) {
//	s.mutex.RLock()
//	defer s.mutex.RUnlock()
//
//	return s.StreamByID(streamID), ErrorStreamNotFound()
//}
//
//func (s *Session) RunIFNotRunning(streamID string) {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//
//	if stream, ok := s.Streams[streamID]; ok {
//		if stream.OnDemand && !stream.RunLock {
//			stream.RunLock = true
//
//			go s.LooperFn(stream)
//		}
//	}
//}
//
//func (s *Session) RunUnlock(uuid string) {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//
//	if stream, ok := s.Streams[uuid]; ok {
//		if stream.OnDemand && stream.RunLock {
//			stream.RunLock = false
//
//			s.Streams[uuid] = stream
//		}
//	}
//}
//
//func (s *Session) StreamExits(steamID string) (ok bool) {
//	s.mutex.RLock()
//	defer s.mutex.RUnlock()
//
//	_, ok = s.Streams[steamID]
//
//	return
//}
//
//func (s *Session) StreamHasChannel(streamID, channelID string) (ok bool) {
//	s.mutex.RLock()
//	defer s.mutex.RUnlock()
//
//	stream, ok := s.Streams[streamID]
//	if !ok {
//		return
//	}
//
//	return ok && stream.HasChannel(channelID)
//}
//
//func (s *Session) ChannelHasSubscriber(uuid string) bool {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//	stream, ok := s.Streams[uuid]
//
//	return ok && stream.HasSubscriber()
//}
//
//func LoadData() (session *Session) {
//	data, err := ioutil.ReadFile("config.json")
//	if err == nil {
//		err = json.Unmarshal(data, &session)
//		if err != nil {
//			log.Fatalln(err)
//		}
//		for i, v := range session.Streams {
//			v.Channels = make(map[string]Subscriber)
//			session.Streams[i] = v
//		}
//	} else {
//		addr := flag.String("listen", "8083", "HTTP host:port")
//		udpMin := flag.Int("udp_min", 0, "WebRTC UDP port min")
//		udpMax := flag.Int("udp_max", 0, "WebRTC UDP port max")
//		iceServer := flag.String("ice_server", "", "ICE Server")
//		flag.Parse()
//
//		session.Server.HTTPPort = *addr
//		session.Server.WebRTCPortMin = uint16(*udpMin)
//		session.Server.WebRTCPortMax = uint16(*udpMax)
//		if len(*iceServer) > 0 {
//			session.Server.ICEServers = []string{*iceServer}
//		}
//
//		session.Streams = make(map[string]Stream)
//	}
//	return session
//}
//
//func (s *Session) Cast(uuid string, Packet av.Packet) {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//
//	for _, v := range s.Streams[uuid].Channels {
//		if len(v.packets) < cap(v.packets) {
//			v.packets <- Packet
//		}
//	}
//}
//
//func (s *Session) RegisterTrack(steamID string, tracks []av.CodecData) bool {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//
//	stream, ok := s.Streams[steamID]
//	if ok {
//		stream.Tracks = tracks
//		s.Streams[steamID] = stream
//
//		return true
//	}
//
//	return false
//}
//
//func (s *Session) StreamTracks(steamID string) []av.CodecData {
//	for i := 0; i < 100; i++ {
//		s.mutex.RLock()
//		stream, ok := s.Streams[steamID]
//		s.mutex.RUnlock()
//
//		if !ok {
//			return nil
//		}
//
//		if stream.HasH264Codecs() {
//			if ok = stream.CheckH264Codecs(); !ok {
//				break
//			}
//		}
//
//		return stream.Tracks
//	}
//
//	time.Sleep(50 * time.Millisecond)
//
//	return nil
//}
//
//func (s *Session) StreamList() (first string, streams []string) {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//
//	for stream := range s.Streams {
//		if len(first) == 0 {
//			first = stream
//		}
//		streams = append(streams, stream)
//	}
//
//	return
//}
//
//func (s *Session) StreamByID(streamID string) *Stream {
//	if stream, ok := s.Streams[streamID]; ok {
//		return &stream
//	}
//
//	return nil
//}
//
//func (st *Stream) CheckH264Codecs() bool {
//	if st.Tracks == nil {
//		return false
//	}
//
//	// TODO Delete test
//	for _, codec := range st.Tracks {
//		if codec.Type() == av.H264 {
//			codecVideo := codec.(h264parser.CodecData)
//			if codecVideo.SPS() == nil || codecVideo.PPS() == nil || codecVideo.EmptySPS() || codecVideo.EmptyPPS() {
//				log.Println("Bad Video CheckH264Codecs SPS or PPS Wait")
//
//				time.Sleep(50 * time.Millisecond)
//				continue
//			}
//		}
//	}
//
//	return true
//}
//
//func (st *Stream) HasH264Codecs() bool {
//	if st.Tracks == nil {
//		return false
//	}
//
//	for _, codec := range st.Tracks {
//		if codec.Type() == av.H264 {
//			return true
//		}
//	}
//
//	return false
//}
