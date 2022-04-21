package session

import (
	"log"
	"time"

	"github.com/teocci/go-stream-av/av"
)

// AddChannel add stream
func (s *Session) AddChannel(streamID, channelID string, channel Channel) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.Streams[streamID]; !ok {
		return ErrorStreamNotFound()
	}

	if ok := s.Streams[streamID].HasChannel(channelID); ok {
		return ErrorChannelAlreadyExists()
	}

	channel.InitChannel()
	s.Streams[streamID].Channels[channelID] = channel
	if !channel.OnDemand {
		channel.RunLock = true
		go StreamServerRunStreamDo(streamID, channelID)
	}

	err := s.SaveConfig()

	if err != nil {
		return err
	}

	return nil
}

// EditChannel edit stream
func (s *Session) EditChannel(streamID, channelID string, newChannel Channel) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, ok := s.Streams[streamID]
	if !ok {
		return ErrorStreamNotFound()
	}

	channel, ok := stream.Channels[channelID]
	if !ok {
		return ErrorChannelNotFound()
	}

	if channel.RunLock {
		channel.Signals <- SignalStreamStopped
	}

	newChannel.InitChannel()
	s.Streams[streamID].Channels[channelID] = newChannel
	if !newChannel.OnDemand {
		newChannel.RunLock = true
		go StreamServerRunStreamDo(streamID, channelID)
	}

	err := s.SaveConfig()
	if err != nil {
		return err
	}

	return nil
}

// DeleteChannel stream
func (s *Session) DeleteChannel(streamID, channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, ok := s.Streams[streamID]
	if !ok {
		return ErrorStreamNotFound()
	}

	channel, ok := stream.Channels[channelID]
	if !ok {
		return ErrorChannelNotFound()
	}

	if channel.RunLock {
		channel.Signals <- SignalStreamStopped
	}

	delete(s.Streams[streamID].Channels, channelID)

	err := s.SaveConfig()
	if err != nil {
		return err
	}

	return nil
}

// RunAllChannels run all stream go
func (s *Session) RunAllChannels() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for streamID, stream := range s.Streams {
		for channelID, channel := range stream.Channels {
			if !channel.OnDemand {
				channel.RunLock = true

				go StreamServerRunStreamDo(streamID, channelID)

				stream.Channels[channelID] = channel
				s.Streams[streamID] = stream
			}
		}
	}
}

// RunChannel one stream and lock
func (s *Session) RunChannel(streamID, channelID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			if !channel.RunLock {
				channel.RunLock = true

				go StreamServerRunStreamDo(streamID, channelID)

				stream.Channels[channelID] = channel
				s.Streams[streamID] = stream
			}
		}
	}
}

// UnlockChannel unlock status to no lock
func (s *Session) UnlockChannel(streamID, channelID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			channel.RunLock = false

			stream.Channels[channelID] = channel
			s.Streams[streamID] = stream
		}
	}
}

// GetChannel get stream channel
func (s *Session) GetChannel(streamID, channelID string) (*Channel, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, ok := s.Streams[streamID]
	if !ok {
		return nil, ErrorStreamNotFound()
	}

	channel, ok := stream.Channels[channelID]
	if !ok {
		return nil, ErrorChannelNotFound()
	}

	return &channel, nil
}

// ChannelExist check stream exist
func (s *Session) ChannelExist(streamID, channelID string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			channel.ACK = time.Now()
			stream.Channels[channelID] = channel
			s.Streams[streamID] = stream

			return true
		}
	}

	return false
}

// ReloadChannel reload stream
func (s *Session) ReloadChannel(streamID, channelID string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {

			channel.Signals <- SignalStreamRestarted

			return nil
		}
	}

	return ErrorStreamNotFound()
}

// ChannelInfo return stream info
func (s *Session) ChannelInfo(streamID, channelID string) (*Channel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			return &channel, nil
		}
	}

	return nil, ErrorStreamNotFound()
}

// ChannelCodecs get stream codec storage or wait
func (s *Session) ChannelCodecs(streamID, channelID string) ([]av.CodecData, error) {
	for i := 0; i < 100; i++ {
		s.mutex.RLock()
		stream, ok := s.Streams[streamID]
		s.mutex.RUnlock()

		if !ok {
			return nil, ErrorStreamNotFound()
		}

		channel, ok := stream.Channels[channelID]
		if !ok {
			return nil, ErrorChannelNotFound()
		}

		if channel.Tracks != nil {
			return channel.Tracks, nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	return nil, ErrorStreamChannelCodecNotFound()
}

// UpdateChannelStatus change stream status
func (s *Session) UpdateChannelStatus(streamID, channelID string, status int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			channel.Status = status
			stream.Channels[channelID] = channel
			s.Streams[streamID] = stream
		}
	}
}

// ChannelCast broadcast stream
func (s *Session) ChannelCast(streamID string, channelID string, packet *av.Packet) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			if channel.HasSubscribers() {
				for _, subscriber := range channel.Subscribers {
					if subscriber.Mode == RTSP {
						continue
					}
					if len(subscriber.Packet) < 1000 {
						subscriber.Packet <- packet
					} else if len(subscriber.Signals) < 10 {
						// send stop Signals to client
						subscriber.Signals <- SignalStreamStopped
						// No need close Socket only send signal to reader / writer Socket closed if client go to offline
					}
				}

				channel.ACK = time.Now()
				stream.Channels[channelID] = channel
				s.Streams[streamID] = stream
			}
		}
	}
}

// ChannelCastProxy broadcast stream
func (s *Session) ChannelCastProxy(streamID string, channelID string, rtp *[]byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			if channel.HasSubscribers() {
				for _, subscriber := range channel.Subscribers {
					if subscriber.Mode != RTSP {
						continue
					}

					if len(subscriber.RTP) < 1000 {
						subscriber.RTP <- rtp
					} else if len(subscriber.Signals) < 10 {
						// send stop Signals to client
						subscriber.Signals <- SignalStreamStopped
						err := subscriber.Socket.Close()
						if err != nil {
							log.Printf("CastProxy [stream][%s] | [channel][%s] | Error: %s\n", streamID, channelID, err)
						}
					}
				}

				channel.ACK = time.Now()

				stream.Channels[channelID] = channel
				s.Streams[streamID] = stream
			}
		}
	}
}

// UpdateCodecInfo update stream codec information
func (s *Session) UpdateCodecInfo(streamID string, channelID string, tracks []av.CodecData, sdp []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			channel.Tracks = tracks
			channel.SDP = sdp

			stream.Channels[channelID] = channel
			s.Streams[streamID] = stream
		}
	}
}

// ChannelSDP codec storage or wait
func (s *Session) ChannelSDP(streamID string, channelID string) ([]byte, error) {
	for i := 0; i < 100; i++ {
		s.mutex.RLock()
		stream, ok := s.Streams[streamID]
		s.mutex.RUnlock()

		if !ok {
			return nil, ErrorStreamNotFound()
		}

		channel, ok := stream.Channels[channelID]
		if !ok {
			return nil, ErrorChannelNotFound()
		}

		if len(channel.SDP) > 0 {
			return channel.SDP, nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	return nil, ErrorEmptySDP()
}
