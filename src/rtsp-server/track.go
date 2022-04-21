// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-07
package rtsp_server

import (
	"fmt"
	"strconv"
	"strings"

	psdp "github.com/pion/sdp/v3"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/base"
)

// Track is a RTSP track.
type Track interface {
	// ClockRate returns the track clock rate.
	ClockRate() int
	// GetControl returns the track control.
	GetControl() string
	// SetControl sets the track control.
	SetControl(string)
	// MediaDescription returns the track media description in SDP format.
	MediaDescription() *psdp.MediaDescription
	clone() Track
	url(*base.URL) (*base.URL, error)
}

func newTrackFromMediaDescription(md *psdp.MediaDescription) (Track, error) {
	control := func() string {
		for _, attr := range md.Attributes {
			if attr.Key == "control" {
				return attr.Value
			}
		}
		return ""
	}()

	encodingCR, payloadType := func() (string, uint8) {
		rtpmap, ok := md.Attribute(mediaDescriptorRTPMapKey)
		if !ok {
			return "", 0
		}
		rtpmap = strings.TrimSpace(rtpmap)

		attributes := strings.Split(rtpmap, " ")
		if len(attributes) != 2 {
			return "", 0
		}

		tmp, err := strconv.ParseInt(attributes[0], 10, 64)
		if err != nil {
			return "", 0
		}

		payloadType := uint8(tmp)

		return attributes[1], payloadType
	}()

	switch {
	case md.MediaName.Media == "video":
		if encodingCR == "H264/90000" {
			return newTrackH264FromMediaDescription(control, payloadType, md)
		}

	case md.MediaName.Media == "audio":
		switch {
		case len(md.MediaName.Formats) == 1 && md.MediaName.Formats[0] == "0":
			return newTrackPCMUFromMediaDescription(control, encodingCR, md)

		case strings.HasPrefix(strings.ToLower(encodingCR), "mpeg4-generic/"):
			return newTrackAACFromMediaDescription(control, payloadType, md)

		case strings.HasPrefix(encodingCR, "opus/"):
			return newTrackOpusFromMediaDescription(control, payloadType, encodingCR, md)
		}
	}

	return newTrackGenericFromMediaDescription(control, md)
}

type trackBase struct {
	control string
}

// GetControl gets the track control.
func (t *trackBase) GetControl() string {
	return t.control
}

// SetControl sets the track control.
func (t *trackBase) SetControl(c string) {
	t.control = c
}

func (t *trackBase) url(contentBase *base.URL) (*base.URL, error) {
	if contentBase == nil {
		return nil, fmt.Errorf("Content-Base header not provided")
	}

	control := t.GetControl()

	// no control attribute, use base URL
	if control == "" {
		return contentBase, nil
	}

	// control attribute contains an absolute path
	if strings.HasPrefix(control, "rtsp://") {
		ur, err := base.ParseURL(control)
		if err != nil {
			return nil, err
		}

		// copy host and credentials
		ur.Host = contentBase.Host
		ur.User = contentBase.User
		return ur, nil
	}

	// control attribute contains a relative control attribute
	// insert the control attribute at the end of the URL
	// if there's a query, insert it after the query
	// otherwise insert it after the path
	strURL := contentBase.String()
	if control[0] != '?' && !strings.HasSuffix(strURL, "/") {
		strURL += "/"
	}

	ur, _ := base.ParseURL(strURL + control)
	return ur, nil
}
