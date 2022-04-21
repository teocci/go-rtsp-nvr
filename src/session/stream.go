package session

//Stream struct
type Stream struct {
	ID   string `json:"-"`
	Name string `json:"name"`
	//URL            string `json:"url"`
	//Status         bool   `json:"status"`
	//OnDemand       bool   `json:"on_demand"`
	//DisableAudio   bool   `json:"disable_audio"`
	//Debug          bool   `json:"debug"`
	//RunLock        bool   `json:"-"`
	//Tracks         []av.CodecData
	Channels map[string]Channel `json:"channels"`
}

func (st *Stream) InitChannels(streamID string) {
	for channelID, channel := range st.Channels {
		channel.InitChannel()
		if !channel.OnDemand {
			channel.RunLock = true
			st.Channels[channelID] = channel
			go StreamServerRunStreamDo(streamID, channelID)
		} else {
			st.Channels[channelID] = channel
		}
	}
}

func (st Stream) HasChannel(channelID string) (ok bool) {
	_, ok = st.Channels[channelID]

	return
}

func (st *Stream) Channel(channelID string) *Channel {
	if channel, ok := st.Channels[channelID]; ok {
		return &channel
	}

	return nil
}

func (st *Stream) StopChannels() {
	for _, channel := range st.Channels {
		channel.End()
	}
}

func (st *Stream) RestartChannels() {
	for _, channel := range st.Channels {
		channel.Restart()
	}
}

func (st *Stream) LoadChannels() {
	for channelID, channel := range st.Channels {
		st.Channels[channelID] = channel
	}
}
