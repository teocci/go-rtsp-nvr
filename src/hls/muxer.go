package hls

import (
	"context"
	"log"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/teocci/go-stream-av/av"
)

// Muxer struct
type Muxer struct {
	mutex          sync.RWMutex
	UUID           string             // Current UUID
	MSN            int                // Current Media Sequence Number
	FPS            int                // Current FPS
	MediaSequence  int                // Current MediaSequence
	CacheM3U8      string             // Current index cache
	Segment        *Segment           // Current segment link
	Segments       map[int]*Segment   // Current segments group
	FragmentID     int                // Current fragment id
	FragmentCtx    context.Context    // chan 1-N
	FragmentCancel context.CancelFunc // chan 1-N
}

// NewMuxer Segments
func NewMuxer(uuid string) *Muxer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Muxer{
		UUID:           uuid,
		MSN:            -1,
		Segments:       make(map[int]*Segment),
		FragmentCtx:    ctx,
		FragmentCancel: cancel,
	}
}

// SetFPS func
func (m *Muxer) SetFPS(fps int) {
	m.FPS = fps
}

// WritePacket func
func (m *Muxer) WritePacket(packet *av.Packet) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// TODO delete packet.IsKeyFrame if need no EXT-X-INDEPENDENT-SEGMENTS
	if packet.IsKeyFrame && (m.Segment == nil || m.Segment.GetDuration().Seconds() >= 4) {
		if m.Segment != nil {
			m.Segment.Close()
			if len(m.Segments) > 6 {
				delete(m.Segments, m.MSN-6)
				m.MediaSequence++
			}
		}
		m.Segment = m.NewSegment()
		m.Segment.SetFPS(m.FPS)
	}

	m.Segment.WritePacket(packet)
	CurrentFragmentID := m.Segment.GetFragmentID()
	if CurrentFragmentID != m.FragmentID {
		m.UpdateIndexM3u8()
	}
	m.FragmentID = CurrentFragmentID
}

// UpdateIndexM3u8 func
func (m *Muxer) UpdateIndexM3u8() {
	var header string
	var body string
	var partTarget time.Duration
	var segmentTarget = time.Second * 2

	for _, segmentKey := range m.SortSegments(m.Segments) {
		for _, fragmentKey := range m.SortFragment(m.Segments[segmentKey].Fragments) {
			if m.Segments[segmentKey].Fragments[fragmentKey].Finish {
				var independent string
				if m.Segments[segmentKey].Fragments[fragmentKey].Independent {
					independent = ",INDEPENDENT=YES"
				}
				body += "#EXT-X-PART:DURATION=" + strconv.FormatFloat(m.Segments[segmentKey].Fragments[fragmentKey].GetDuration().Seconds(), 'f', 5, 64) + "" + independent + ",URI=\"fragment/" + strconv.Itoa(segmentKey) + "/" + strconv.Itoa(fragmentKey) + "/0qrm9ru6." + strconv.Itoa(fragmentKey) + ".m4s\"\n"
				partTarget = m.Segments[segmentKey].Fragments[fragmentKey].Duration
			} else {
				body += "#EXT-X-PRELOAD-HINT:TYPE=PART,URI=\"fragment/" + strconv.Itoa(segmentKey) + "/" + strconv.Itoa(fragmentKey) + "/0qrm9ru6." + strconv.Itoa(fragmentKey) + ".m4s\"\n"
			}
		}
		if m.Segments[segmentKey].Finish {
			segmentTarget = m.Segments[segmentKey].Duration
			body += "#EXT-X-PROGRAM-DATE-TIME:" + m.Segments[segmentKey].Time.Format("2006-01-02T15:04:05.000000Z") + "\n#EXTINF:" + strconv.FormatFloat(m.Segments[segmentKey].Duration.Seconds(), 'f', 5, 64) + ",\n"
			body += "segment/" + strconv.Itoa(segmentKey) + "/" + m.UUID + "." + strconv.Itoa(segmentKey) + ".m4s\n"
		}
	}
	header += "#EXTM3U\n"
	header += "#EXT-X-TARGETDURATION:" + strconv.Itoa(int(math.Round(segmentTarget.Seconds()))) + "\n"
	header += "#EXT-X-VERSION:7\n"
	header += "#EXT-X-INDEPENDENT-SEGMENTS\n"
	header += "#EXT-X-SERVER-CONTROL:CAN-BLOCK-RELOAD=YES,PART-HOLD-BACK=" + strconv.FormatFloat(partTarget.Seconds()*4, 'f', 5, 64) + ",HOLD-BACK=" + strconv.FormatFloat(segmentTarget.Seconds()*4, 'f', 5, 64) + "\n"
	header += "#EXT-X-MAP:URI=\"init.mp4\"\n"
	header += "#EXT-X-PART-INF:PART-TARGET=" + strconv.FormatFloat(partTarget.Seconds(), 'f', 5, 64) + "\n"
	header += "#EXT-X-MEDIA-SEQUENCE:" + strconv.Itoa(m.MediaSequence) + "\n"
	header += body
	m.CacheM3U8 = header
	m.PlaylistUpdate()
}

// PlaylistUpdate func
func (m *Muxer) PlaylistUpdate() {
	m.FragmentCancel()
	m.FragmentCtx, m.FragmentCancel = context.WithCancel(context.Background())
}

// GetSegment func
func (m *Muxer) GetSegment(segment int) (p []*av.Packet, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	err = ErrorStreamNotFound()
	if segmentTmp, ok := m.Segments[segment]; ok && len(segmentTmp.Fragments) > 0 {
		var res []*av.Packet
		for _, v := range m.SortFragment(segmentTmp.Fragments) {
			res = append(res, segmentTmp.Fragments[v].Packets...)
		}

		return res, nil
	}

	return nil, err
}

// GetFragment func
func (m *Muxer) GetFragment(segment int, fragment int) ([]*av.Packet, error) {
	m.mutex.Lock()
	if segmentTmp, segmentTmpOK := m.Segments[segment]; segmentTmpOK {
		if fragmentTmp, fragmentTmpOK := segmentTmp.Fragments[fragment]; fragmentTmpOK {
			if fragmentTmp.Finish {
				m.mutex.Unlock()

				return fragmentTmp.Packets, nil
			} else {
				m.mutex.Unlock()
				pck, err := m.WaitFragment(time.Second*1, segment, fragment)
				if err != nil {
					return nil, err
				}

				return pck, err
			}
		}
	}
	m.mutex.Unlock()

	return nil, ErrorStreamNotFound()
}

// GetIndexM3u8 func
func (m *Muxer) GetIndexM3u8(needMSN int, needPart int) (string, error) {
	m.mutex.Lock()

	if len(m.CacheM3U8) != 0 && ((needMSN == -1 || needPart == -1) || (needMSN-m.MSN > 1) || (needMSN == m.MSN && needPart < m.FragmentID)) {
		m.mutex.Unlock()

		return m.CacheM3U8, nil
	} else {
		m.mutex.Unlock()

		index, err := m.WaitIndex(time.Second*3, needMSN, needPart)
		if err != nil {
			return "", err
		}

		return index, err
	}
}

// WaitFragment func
func (m *Muxer) WaitFragment(timeOut time.Duration, segment, fragment int) (p []*av.Packet, err error) {
	p, err = nil, ErrorStreamNotFound()
	select {
	case <-time.After(timeOut):
		return
	case <-m.FragmentCtx.Done():
		m.mutex.Lock()
		defer m.mutex.Unlock()
		if segmentTmp, segmentTmpOK := m.Segments[segment]; segmentTmpOK {
			if fragmentTmp, fragmentTmpOK := segmentTmp.Fragments[fragment]; fragmentTmpOK {
				if fragmentTmp.Finish {
					return fragmentTmp.Packets, nil
				}
			}
		}

		return
	}
}

// WaitIndex func
func (m *Muxer) WaitIndex(timeOut time.Duration, segment, fragment int) (string, error) {
	for {
		select {
		case <-time.After(timeOut):
			return "", ErrorStreamNotFound()
		case <-m.FragmentCtx.Done():
			m.mutex.Lock()
			if m.MSN < segment || (m.MSN == segment && m.FragmentID < fragment) {
				log.Println("wait req", m.MSN, m.FragmentID, segment, fragment)
				m.mutex.Unlock()
				continue
			}
			m.mutex.Unlock()
			return m.CacheM3U8, nil
		}
	}
}

// SortFragment func
func (m *Muxer) SortFragment(val map[int]*Fragment) []int {
	keys := make([]int, len(val))
	i := 0
	for k := range val {
		keys[i] = k
		i++
	}
	sort.Ints(keys)
	return keys
}

// SortSegments fuc
func (m *Muxer) SortSegments(val map[int]*Segment) []int {
	keys := make([]int, len(val))
	i := 0
	for k := range val {
		keys[i] = k
		i++
	}
	sort.Ints(keys)
	return keys
}

func (m *Muxer) Close() {

}
