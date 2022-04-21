// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-11
package rtsp_server

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"

	psdp "github.com/pion/sdp/v3"
)

func trackH264GetSPSPPS(md *psdp.MediaDescription) ([]byte, []byte, error) {
	v, ok := md.Attribute("fmtp")
	if !ok {
		return nil, nil, ErrorFMTPAttributeIsMissing()
	}

	aList := strings.SplitN(v, " ", 2)
	if len(aList) != 2 {
		return nil, nil, ErrorInvalidFMTPAttribute(v)
	}

	for _, kv := range strings.Split(aList[1], ";") {
		kv = strings.Trim(kv, " ")

		if len(kv) == 0 {
			continue
		}

		pList := strings.SplitN(kv, "=", 2)
		if len(pList) != 2 {
			return nil, nil, ErrorInvalidFMTPAttribute(v)
		}

		if pList[0] == keyParameterSets {
			params := strings.Split(pList[1], ",")
			if len(params) < 2 {
				return nil, nil, ErrorInvalidParameterSets(v)
			}

			sps, err := base64.StdEncoding.DecodeString(params[0])
			if err != nil {
				return nil, nil, ErrorInvalidParameterSets(v)
			}

			pps, err := base64.StdEncoding.DecodeString(params[1])
			if err != nil {
				return nil, nil, ErrorInvalidParameterSets(v)
			}

			return sps, pps, nil
		}
	}

	return nil, nil, ErrorParameterSetsAreMissing(v)
}

// TrackH264 is a H264 track.
type TrackH264 struct {
	trackBase
	payloadType uint8
	sps         []byte
	pps         []byte
	extraData   []byte
	mutex       sync.RWMutex
}

// NewTrackH264 allocates a TrackH264.
func NewTrackH264(payloadType uint8, sps []byte, pps []byte, extra []byte) (*TrackH264, error) {
	return &TrackH264{
		payloadType: payloadType,
		sps:         sps,
		pps:         pps,
		extraData:   extra,
	}, nil
}

func newTrackH264FromMediaDescription(control string, payloadType uint8, md *psdp.MediaDescription) (t *TrackH264, err error) {
	var sps, pps []byte

	t = &TrackH264{
		trackBase: trackBase{
			control: control,
		},
		payloadType: payloadType,
	}

	sps, pps, err = trackH264GetSPSPPS(md)
	if err == nil {
		t.sps = sps
		t.pps = pps
	}

	return t, nil
}

// ClockRate returns the track clock rate.
func (t *TrackH264) ClockRate() int {
	return 90000
}

func (t *TrackH264) clone() Track {
	return &TrackH264{
		trackBase:   t.trackBase,
		payloadType: t.payloadType,
		sps:         t.sps,
		pps:         t.pps,
		extraData:   t.extraData,
	}
}

// SPS returns the track SPS.
func (t *TrackH264) SPS() []byte {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.sps
}

// PPS returns the track PPS.
func (t *TrackH264) PPS() []byte {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.pps
}

// ExtraData returns the track extra data.
func (t *TrackH264) ExtraData() []byte {
	return t.extraData
}

// SetSPS sets the track SPS.
func (t *TrackH264) SetSPS(v []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.sps = v
}

// SetPPS sets the track PPS.
func (t *TrackH264) SetPPS(v []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.pps = v
}

func (t *TrackH264) payloadTypeInt() string {
	return strconv.FormatInt(int64(t.payloadType), 10)
}

func (t *TrackH264) profileLevelId() string {
	return strings.ToUpper(hex.EncodeToString(t.sps[1:4]))
}

const (
	formatParameterAppender = "%s; %s"
	formatPayloadPacketMode = "%s packetization-mode=1"
	formatParameterSets     = "sprop-parameter-sets=%s"
	formatProfileLevelID    = "profile-level-id=%s"

	// SDP Attributes
	// a=rtpmap:<payload type> <encoding name>/<clock rate> [/<encoding parameters>]
	formatMDAttributeRTPMapValue = "%s H264/90000"
)

func appendMDParameter(m, s string) string {
	return fmt.Sprintf(formatParameterAppender, m, s)
}

// MediaDescription returns the track media description in SDP format.
func (t *TrackH264) MediaDescription() *psdp.MediaDescription {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	pType := t.payloadTypeInt()
	rtpMap := fmt.Sprintf(formatMDAttributeRTPMapValue, pType)

	fmtp := fmt.Sprintf(formatPayloadPacketMode, pType)

	var params []string
	if t.sps != nil {
		params = append(params, base64.StdEncoding.EncodeToString(t.sps))
	}
	if t.pps != nil {
		params = append(params, base64.StdEncoding.EncodeToString(t.pps))
	}
	if t.extraData != nil {
		params = append(params, base64.StdEncoding.EncodeToString(t.extraData))
	}
	fmtp = appendMDParameter(fmtp, fmt.Sprintf(formatParameterSets, strings.Join(params, ",")))

	if len(t.sps) >= 4 {
		fmtp = appendMDParameter(fmtp, fmt.Sprintf(formatProfileLevelID, t.profileLevelId()))
	}

	return &psdp.MediaDescription{
		MediaName: psdp.MediaName{
			Media:   "video",
			Protos:  []string{"RTP", "AVP"},
			Formats: []string{pType},
		},
		Attributes: []psdp.Attribute{
			{
				Key:   "rtpmap",
				Value: rtpMap,
			},
			{
				Key:   "fmtp",
				Value: fmtp,
			},
			{
				Key:   "control",
				Value: t.control,
			},
		},
	}
}
