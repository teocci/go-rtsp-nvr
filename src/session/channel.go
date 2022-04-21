package session

import (
	"time"

	"github.com/teocci/go-rtsp-nvr/src/hls"
	"github.com/teocci/go-stream-av/av"
)

// Channel struct
type Channel struct {
	ID                 string `json:"-"`
	StreamID           string `json:"-"`
	Name               string `json:"name,omitempty"`
	URL                string `json:"url"`
	Status             int    `json:"status"`
	OnDemand           bool   `json:"on_demand"`
	Audio              bool   `json:"audio"`
	Debug              bool   `json:"debug"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify,omitempty"`
	RunLock            bool   `json:"-"`
	SDP                []byte
	ACK                time.Time
	Signals            chan int
	HLSPublisher       HLSPublisher
	Tracks             []av.CodecData
	Subscribers        map[string]*Subscriber
}

func (ch *Channel) InitChannel() {
	ch.ACK = time.Now().Add(-255 * time.Hour)
	ch.HLSPublisher = NewHLSPublisher()
	ch.Signals = make(chan int, 100)
	ch.Subscribers = make(map[string]*Subscriber)

	return
}

func (ch *Channel) HasSubscribers() bool {
	return len(ch.Subscribers) > 0
}

func (ch *Channel) HasSubscriber(sid string) (ok bool) {
	_, ok = ch.Subscribers[sid]

	return
}

// AddSubscriber Add New Client to Translations
func (ch *Channel) AddSubscriber(mode SubscriberMode) (subscriber *Subscriber) {
	subscriber = NewSubscriber(mode)

	ch.Subscribers[subscriber.ID] = subscriber
	ch.ACK = time.Now()

	return
}

// RemoveSubscriber Delete Client
func (ch *Channel) RemoveSubscriber(sid string) {
	delete(ch.Subscribers, sid)
}

func (ch *Channel) End() {
	if ch.RunLock {
		ch.Signals <- SignalStreamStopped
	}
}

func (ch *Channel) Restart() {
	if ch.RunLock {
		ch.Signals <- SignalStreamRestarted
	}
}

func (ch *Channel) InitHLSMuxer(uuid string) {
	ch.HLSPublisher.Muxer = hls.NewMuxer(uuid)
}

func (ch Channel) WasRecentlyACK() bool {
	return !(time.Now().Sub(ch.ACK).Seconds() > 30)
}
