package session

// HasSubscriber check is client ext
func (s *Session) HasSubscriber(streamID, channelID string) (ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, ok := s.Streams[streamID]
	if !ok {
		return
	}

	channel, ok := stream.Channels[channelID]
	if !ok {
		return
	}

	return channel.WasRecentlyACK()
}

// AddSubscriber Add New Client to Translations
func (s *Session) AddSubscriber(streamID, channelID string, mode SubscriberMode) (subscriber *Subscriber, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, ok := s.Streams[streamID]
	if !ok {
		return nil, ErrorStreamNotFound()
	}

	channel, ok := s.Streams[streamID].Channels[channelID]
	if !ok {
		return nil, ErrorChannelNotFound()
	}

	channel.AddSubscriber(mode)
	stream.Channels[channelID] = channel
	s.Streams[streamID] = stream

	return subscriber, nil
}

// RemoveSubscriber Delete Client
func (s *Session) RemoveSubscriber(streamID, channelID, sid string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.Streams[streamID]; !ok {
		return
	}

	if _, ok := s.Streams[streamID].Channels[channelID]; !ok {
		delete(s.Streams[streamID].Channels[channelID].Subscribers, sid)
	}
}
