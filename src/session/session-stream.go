// Package session
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-01
package session

// AddStream add stream
func (s *Session) AddStream(streamID string, stream Stream) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	//TODO create empty map bug save https://github.com/liip/sheriff empty not nil map[] != {} json
	//data, err := sheriff.Marshal(&sheriff.Options{
	//		Groups:     []string{"config"},
	//		ApiVersion: v2,
	//	}, s)
	//Not Work map[] != {}
	s.InitNilStreams()

	if _, ok := s.Streams[streamID]; ok {
		return ErrorStreamAlreadyExists()
	}

	stream.InitChannels(streamID)

	s.Streams[streamID] = stream

	err := s.SaveConfig()
	if err != nil {
		return err
	}

	return nil
}

// EditStream edit stream
func (s *Session) EditStream(streamID string, stream Stream) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if currentStream, ok := s.Streams[streamID]; ok {
		for channelID, channel := range currentStream.Channels {
			if channel.RunLock {
				currentStream.Channels[channelID] = channel
				s.Streams[streamID] = currentStream

				channel.Signals <- SignalStreamStopped
			}
		}

		stream.InitChannels(streamID)

		s.Streams[streamID] = stream

		err := s.SaveConfig()
		if err != nil {
			return err
		}

		return nil
	}

	return ErrorStreamNotFound()
}
