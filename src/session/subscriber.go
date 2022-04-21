package session

import (
	"github.com/teocci/go-stream-av/av"
	"net"
)

type Subscriber struct {
	ID      string
	Mode    SubscriberMode
	Signals chan int
	Packet  chan *av.Packet
	RTP     chan *[]byte
	Socket  net.Conn
}

func NewSubscriber(mode SubscriberMode) *Subscriber {
	return &Subscriber{
		ID:      generateUUID(),
		Mode:    mode,
		Packet:  make(chan *av.Packet, 2000),
		RTP:     make(chan *[]byte, 2000),
		Signals: make(chan int, 100),
	}
}
