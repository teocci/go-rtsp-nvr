// Package session
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-01
package session

import (
	"time"

	"github.com/teocci/go-stream-av/av"
)

// NewHLSMuxer new muxer init
func (s *Session) NewHLSMuxer(streamID, channelID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			channel.InitHLSMuxer(streamID)
			stream.Channels[channelID] = channel
			s.Streams[streamID] = stream
		}
	}
}

// HLSMuxerSetFPS write Packet
func (s *Session) HLSMuxerSetFPS(uuid string, channelID string, fps int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[uuid]; ok {
		if channel, ok := stream.Channels[channelID]; ok && channel.HLSPublisher.Muxer != nil {
			channel.HLSPublisher.SetMuxerFPS(fps)
			stream.Channels[channelID] = channel
			s.Streams[uuid] = stream
		}
	}
}

// HLSMuxerWritePacket write Packet
func (s *Session) HLSMuxerWritePacket(uuid string, channelID string, packet *av.Packet) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if stream, ok := s.Streams[uuid]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			channel.HLSPublisher.MuxerWritePacket(packet)
		}
	}
}

// HLSMuxerClose close muxer
func (s *Session) HLSMuxerClose(uuid string, channelID string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if stream, ok := s.Streams[uuid]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			channel.HLSPublisher.Muxer.Close()
		}
	}
}

// HLSMuxerM3U8 get m3u8 list
func (s *Session) HLSMuxerM3U8(uuid string, channelID string, msn, part int) (string, error) {
	s.mutex.Lock()
	stream, ok := s.Streams[uuid]
	s.mutex.Unlock()

	if !ok {
		return "", ErrorStreamNotFound()
	}

	if channel, ok := stream.Channels[channelID]; ok {
		return channel.HLSPublisher.M3U8Index(msn, part)
	}

	return "", ErrorChannelNotFound()
}

// HLSMuxerSegment get segment
func (s *Session) HLSMuxerSegment(streamID, channelID string, segment int) ([]*av.Packet, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			return channel.HLSPublisher.Segment(segment)
		}
	}

	return nil, ErrorChannelNotFound()
}

// HLSMuxerFragment get fragment
func (s *Session) HLSMuxerFragment(streamID string, channelID string, segment, fragment int) ([]*av.Packet, error) {
	s.mutex.Lock()
	stream, ok := s.Streams[streamID]
	s.mutex.Unlock()

	if ok {
		if channel, ok := stream.Channels[channelID]; ok {
			return channel.HLSPublisher.Fragment(segment, fragment)
		}
	}

	return nil, ErrorChannelNotFound()
}

// StreamHLSAdd add hls seq to buffer
func (s *Session) StreamHLSAdd(uuid string, channelID string, packets []*av.Packet, dur time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[uuid]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			channel.HLSPublisher.PackSegment(packets, dur)

			stream.Channels[channelID] = channel
			s.Streams[uuid] = stream
		}
	}
}

// StreamHLSm3u8 get hls m3u8 list
func (s *Session) StreamHLSm3u8(streamID, channelID string) (string, int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			out, count := channel.HLSPublisher.M3U8List()

			return out, count, nil
		}
	}

	return "", 0, ErrorStreamNotFound()
}

// StreamHLSTS send hls segment buffer to clients
func (s *Session) StreamHLSTS(streamID, channelID string, seq int) ([]*av.Packet, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			if segment, ok := channel.HLSPublisher.SegmentBuffer[seq]; ok {
				return segment.Packet, nil
			}
		}
	}

	return nil, ErrorStreamNotFound()
}

// StreamHLSFlush delete hls cache
func (s *Session) StreamHLSFlush(streamID string, channelID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if stream, ok := s.Streams[streamID]; ok {
		if channel, ok := stream.Channels[channelID]; ok {
			channel.HLSPublisher.InitSegment()
			stream.Channels[channelID] = channel
			s.Streams[streamID] = stream
		}
	}
}
