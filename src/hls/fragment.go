package hls

import (
	"time"

	"github.com/teocci/go-stream-av/av"
)

// Fragment struct
type Fragment struct {
	Independent bool          // Fragment have i-frame (key frame)
	Finish      bool          // Fragment ready
	Duration    time.Duration // Fragment duration
	Packets     []*av.Packet  // Packet list
}

// NewFragment open new fragment
func (s *Segment) NewFragment() (fragment *Fragment) {
	fragment = &Fragment{}
	s.Fragments[s.FragmentID] = fragment

	return
}

// GetDuration fragment duration
func (f *Fragment) GetDuration() time.Duration {
	return f.Duration
}

// WritePacket to fragment
func (f *Fragment) WritePacket(packet *av.Packet) {
	f.Duration += packet.Duration
	// Checks if it has a keyframe to mark it as independent
	if packet.IsKeyFrame {
		f.Independent = true
	}
	// Appends packet to slice of packet
	f.Packets = append(f.Packets, packet)
}

// Close fragment block
func (f *Fragment) Close() {
	// TODO add callback func
	f.Finish = true
}
