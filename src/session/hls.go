package session

import (
	"fmt"
	"sort"
	"time"

	"github.com/teocci/go-rtsp-nvr/src/hls"
	"github.com/teocci/go-stream-av/av"
)

const (
	formatHLSM3U8 = "#EXTM3U\r\n#EXT-X-TARGETDURATION:4\r\n#EXT-X-VERSION:4\r\n" +
		"#EXT-X-MEDIA-SEQUENCE:%d\r\n"
	formatHLSInfo = "#EXTINF:%.2f,\r\nsegment/%d/file.ts\r\n"
)

type HLSPublisher struct {
	SegmentNumber int
	SegmentBuffer map[int]HLSSegment
	Muxer         *hls.Muxer
}

//HLSSegment HLS cache section
type HLSSegment struct {
	Duration time.Duration
	Packet   []*av.Packet
}

func NewHLSPublisher() HLSPublisher {
	return HLSPublisher{
		SegmentBuffer: make(map[int]HLSSegment),
	}
}

func (hp *HLSPublisher) InitSegment() {
	hp.SegmentBuffer = make(map[int]HLSSegment)
	hp.SegmentNumber = 0
}

func (hp *HLSPublisher) PackSegment(packets []*av.Packet, dur time.Duration) {
	hp.SegmentNumber++
	hp.SegmentBuffer[hp.SegmentNumber] = HLSSegment{Packet: packets, Duration: dur}
	if len(hp.SegmentBuffer) >= 6 {
		delete(hp.SegmentBuffer, hp.SegmentNumber-6-1)
	}
}

func (hp *HLSPublisher) SetMuxerFPS(fps int) {
	hp.Muxer.SetFPS(fps)
}

func (hp *HLSPublisher) MuxerWritePacket(packet *av.Packet) {
	hp.Muxer.WritePacket(packet)
}

func (hp *HLSPublisher) M3U8Index(msn, part int) (string, error) {
	return hp.Muxer.GetIndexM3u8(msn, part)
}

func (hp *HLSPublisher) Segment(segment int) ([]*av.Packet, error) {
	return hp.Muxer.GetSegment(segment)
}

func (hp *HLSPublisher) Fragment(segment, fragment int) ([]*av.Packet, error) {
	return hp.Muxer.GetFragment(segment, fragment)
}

func (hp *HLSPublisher) M3U8List() (out string, count int) {
	// TODO fix  it
	out += fmt.Sprintf(formatHLSM3U8, hp.SegmentNumber)
	var indexes []int
	for k := range hp.SegmentBuffer {
		indexes = append(indexes, k)
	}

	sort.Ints(indexes)
	for _, i := range indexes {
		count++
		out += fmt.Sprintf(formatHLSInfo, hp.SegmentBuffer[i].Duration.Seconds(), i)
	}

	return
}
