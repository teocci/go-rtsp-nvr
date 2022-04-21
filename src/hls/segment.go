package hls

import (
	"time"

	"github.com/teocci/go-stream-av/av"
)

// Segment struct
type Segment struct {
	FPS        int               // Current fps
	Finish     bool              // Segment ready
	Duration   time.Duration     // Segment duration
	Time       time.Time         // Realtime EXT-X-PROGRAM-DATE-TIME
	FragmentID int               // Fragment ID
	Fragment   *Fragment         // Fragment link
	Fragments  map[int]*Fragment // Fragment list
}

// NewSegment func
func (m *Muxer) NewSegment() (res *Segment) {
	res = &Segment{
		FragmentID: -1, // Default fragment -1
		Fragments:  make(map[int]*Fragment),
	}
	// Increase MSN
	m.MSN++
	m.Segments[m.MSN] = res

	return
}

// GetDuration func
func (s *Segment) GetDuration() time.Duration {
	return s.Duration
}

// SetFPS func
func (s *Segment) SetFPS(fps int) {
	s.FPS = fps
}

// WritePacket writes a packet to a fragment
func (s *Segment) WritePacket(packet *av.Packet) {
	if s.Fragment == nil || s.Fragment.GetDuration().Milliseconds() >= s.FragmentLengthMS(s.FPS) {
		if s.Fragment != nil {
			s.Fragment.Close()
		}
		s.FragmentID++
		s.Fragment = s.NewFragment()
	}

	s.Duration += packet.Duration
	s.Fragment.WritePacket(packet)
}

// GetFragmentID func
func (s *Segment) GetFragmentID() int {
	return s.FragmentID
}

// FragmentLengthMS fragment length in ms
func (s *Segment) FragmentLengthMS(fps int) int64 {
	for i := 6; i >= 1; i-- {
		if fps%i == 0 {
			return int64(1e3 / float64(fps) * float64(i))
		}
	}

	return 100
}

// Close segment
func (s *Segment) Close() {
	s.Finish = true
	if s.Fragment != nil {
		s.Fragment.Close()
	}
}
