// Package rtsp_server
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-15
package rtsp_server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/base"
	"github.com/teocci/go-rtsp-nvr/src/rtsp-server/liberrors"
	"github.com/teocci/go-rtsp-nvr/src/session"
)

var (
	Session *session.Session
)

func Start(s *session.Session) {
	Session = s

	log.Printf("%s [%s]", base.TAG, "Server RTSP start")
	server, err := net.Listen("tcp", session.CoreSession.Server.Web.Port)
	if err != nil {
		log.Printf("%s [Error]: %s", base.TAG, err)
		os.Exit(1)
	}
	defer func() {
		err := server.Close()
		if err != nil {
			log.Printf("%s [Error]: %s", base.TAG, err)
		}
	}()

	for {
		cxn, err := server.Accept()
		if err != nil {
			log.Printf("%s [Error]: %s", base.TAG, err)
			os.Exit(1)
		}
		go Run(cxn)
	}
}

//Run func
func Run(conn net.Conn) {
	buf := make([]byte, 4096)
	token, streamID, channelID, in, cSeq := "", "", "0", 0, 0
	var playStarted bool
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
		}
	}()

	err := conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
		return
	}

	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
			return
		}

		cSeq, err = parseCSeq(buf[:n])
		if err != nil {
			log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
		}

		option, err := parseOption(buf[:n])
		if err != nil {
			log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
		}

		err = conn.SetDeadline(time.Now().Add(60 * time.Second))
		log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, string(buf[:n]))
		if err != nil {
			log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
			return
		}

		switch option {
		case base.OPTIONS:
			if playStarted {
				err = SendResponse(streamID, channelID, conn, 200, map[string]string{"CSeq": strconv.Itoa(cSeq), "Public": "DESCRIBE, SETUP, TEARDOWN, PLAY"})
				if err != nil {
					return
				}
				continue
			}
			streamID, channelID, token, err = parseStreamChannel(buf[:n])
			if err != nil {
				log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
				return
			}
			if !session.CoreSession.ChannelExist(streamID, channelID) {
				log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, liberrors.ErrorStreamNotFound())
				err = SendResponse(streamID, channelID, conn, 404, map[string]string{"CSeq": strconv.Itoa(cSeq)})
				if err != nil {
					return
				}
				return
			}

			if !session.RemoteAuthorization("RTSP", streamID, channelID, token, conn.RemoteAddr().String()) {
				log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, liberrors.ErrorStreamNotFound())
				err = SendResponse(streamID, channelID, conn, 401, map[string]string{"CSeq": strconv.Itoa(cSeq)})
				if err != nil {
					return
				}
				return
			}

			session.CoreSession.RunChannel(streamID, channelID)
			err = SendResponse(streamID, channelID, conn, 200, map[string]string{"CSeq": strconv.Itoa(cSeq), "Public": "DESCRIBE, SETUP, TEARDOWN, PLAY"})
			if err != nil {
				return
			}
		case base.SETUP:
			if !strings.Contains(string(buf[:n]), "interleaved") {
				err = SendResponse(streamID, channelID, conn, 461, map[string]string{"CSeq": strconv.Itoa(cSeq)})
				if err != nil {
					return
				}
				continue
			}
			err = SendResponse(streamID, channelID, conn, 200, map[string]string{
				"CSeq":        strconv.Itoa(cSeq),
				"User-Agent:": base.UserAgent,
				"Session":     base.Session,
				"Transport":   fmt.Sprintf(base.TransportFormat, in, in+1),
			})
			if err != nil {
				return
			}
			in = in + 2
		case base.DESCRIBE:
			sdp, err := session.CoreSession.ChannelSDP(streamID, channelID)
			if err != nil {
				log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
				return
			}
			err = SendResponse(streamID, channelID, conn, 200, map[string]string{
				"CSeq":           strconv.Itoa(cSeq),
				"User-Agent:":    base.UserAgent,
				"Session":        base.Session,
				"Content-Type":   base.SDPContentType,
				"Content-Length": strconv.Itoa(len(sdp)),
				"sdp":            string(sdp),
			})
			if err != nil {
				return
			}
		case base.PLAY:
			err = SendResponse(streamID, channelID, conn, 200, map[string]string{
				"CSeq":        strconv.Itoa(cSeq),
				"User-Agent:": base.UserAgent,
				"Session":     base.Session,
			})
			if err != nil {
				return
			}
			playStarted = true
			go playClient(streamID, channelID, conn)
		case base.TEARDOWN:
			err = SendResponse(streamID, channelID, conn, 200, map[string]string{
				"CSeq":        strconv.Itoa(cSeq),
				"User-Agent:": base.UserAgent,
				"Session":     base.Session,
			})
			if err != nil {
				return
			}
			return
		default:
			log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, liberrors.ErrorBadOption(option))
		}
	}
}

// playClient play a subscriber
func playClient(streamID string, channelID string, conn net.Conn) {
	subscriber, err := session.CoreSession.AddSubscriber(streamID, channelID, session.RTSP)
	if err != nil {
		log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
		return
	}
	defer func() {
		session.CoreSession.RemoveSubscriber(streamID, channelID, subscriber.ID)
		log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Info]: %s", base.TAG, streamID, channelID, "Client offline")
		err := conn.Close()
		if err != nil {
			log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
		}
	}()

	noVideo := time.NewTimer(10 * time.Second)

	for {
		select {
		case <-noVideo.C:
			return
		case pck := <-subscriber.RTP:
			noVideo.Reset(10 * time.Second)
			_, err := conn.Write(*pck)
			if err != nil {
				log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
				return
			}
		}
	}
}

// SendResponse func
func SendResponse(streamID string, channelID string, conn net.Conn, code base.StatusCode, headers map[string]string) error {
	var sdp string
	buffer := bytes.Buffer{}
	buffer.WriteString(fmt.Sprintf(base.Version+" %d %s\r\n", code, base.StatusMessage(code)))
	for k, v := range headers {
		if k == "sdp" {
			sdp = v
			continue
		}
		buffer.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	buffer.WriteString(fmt.Sprintf("\r\n"))
	buffer.WriteString(sdp)

	log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [response]: %s", base.TAG, streamID, channelID, buffer.String())
	if _, err := conn.Write(buffer.Bytes()); err != nil {
		log.Printf("%s [stream]:[%s] | [channelID]:[%s] | [Error]: %s", base.TAG, streamID, channelID, err)
		return err
	}
	return nil
}

// parseCSeq func
func parseCSeq(buf []byte) (int, error) {
	s := session.StringInBetween(string(buf), "CSeq: ", "\r\n")
	if len(s) > 0 {
		return session.StringToInt(s), nil
	}

	return 0, liberrors.ErrorParsingCSeq()
}

// parseOption func
func parseOption(buf []byte) (string, error) {
	s := strings.Split(string(buf), " ")
	if len(s) > 0 {
		return s[0], nil
	}

	return "", liberrors.ErrorParsingOptionFail(string(buf))
}

// parseStreamChannel func
func parseStreamChannel(buf []byte) (string, string, string, error) {
	var token string

	uri := session.StringInBetween(string(buf), " ", " ")
	u, err := url.Parse(uri)
	if err == nil {
		token = u.Query().Get("token")
		uri = u.Path
	}

	st := strings.Split(uri, "/")

	if len(st) >= 3 {
		return st[1], st[2], token, nil
	}

	return "", "0", token, liberrors.ErrorParsingStreamFail(string(buf))
}
