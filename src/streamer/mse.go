package streamer

import (
	"log"
	"time"

	"golang.org/x/net/websocket"

	"github.com/teocci/go-rtsp-nvr/src/session"
	"github.com/teocci/go-rtsp-nvr/src/webauth"
	"github.com/teocci/go-stream-av/format/mp4f"
)

type MSERequest struct {
	StreamID  string `json:"stream"`
	ChannelID string `json:"channel"`
	Token     string `json:"token"`
}

func HandleStreamMSE(ws *websocket.Conn) {
	defer func() {
		err := ws.Close()
		log.Println("Client Full Exit", err)
	}()

	req := MSERequest{
		StreamID:  ws.Request().FormValue("uuid"),
		ChannelID: ws.Request().FormValue("channel"),
		Token:     ws.Request().FormValue("token"),
	}

	if Session.WebTokenEnable() && !webauth.RemoteAuthorization(webauth.AuthRequest{
		Proto:   "WS",
		Stream:  req.StreamID,
		Channel: req.ChannelID,
		Token:   req.Token,
		IP:      ws.Request().RemoteAddr,
	}, Session.WebTokenBackend()) {
		err := ErrorRemoteAuthorizationFailed()
		log.Printf("RemoteAuthorization [stream][%s], [channel][%s] | [error]: %s", req.StreamID, req.ChannelID, err)

		return
	}

	if !session.CoreSession.ChannelExist(req.StreamID, req.ChannelID) {
		err := ErrorChannelNotFound()
		log.Printf("ChannelExist [stream][%s], [channel][%s] | [error]: %s", req.StreamID, req.ChannelID, err)

		return
	}

	session.CoreSession.RunChannel(req.StreamID, req.ChannelID)

	err := ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		log.Printf("SetWriteDeadline [stream][%s], [channel][%s] | [error]: %s", req.StreamID, req.ChannelID, err)
		return
	}

	subscriber, err := session.CoreSession.AddSubscriber(req.StreamID, req.ChannelID, session.MSE)
	if err != nil {
		log.Printf("AddSubscriber [stream][%s], [channel][%s] | [error]: %s", req.StreamID, req.ChannelID, err)
		return
	}
	defer session.CoreSession.RemoveSubscriber(req.StreamID, req.ChannelID, subscriber.ID)

	codecs, err := session.CoreSession.ChannelCodecs(req.StreamID, req.ChannelID)
	if err != nil {
		log.Printf("AddSubscriber [stream][%s], [channel][%s] | [error]: %s\n", req.StreamID, req.ChannelID, err)
		return
	}

	var muxerMSE = mp4f.NewMuxer(nil)
	err = muxerMSE.WriteHeader(codecs)
	if err != nil {
		log.Printf("muxerMSE.WriteHeader [stream][%s], [channel][%s] | [error]: %s\n", req.StreamID, req.ChannelID, err)
		return
	}

	meta, buffer := muxerMSE.GetInit(codecs)
	err = websocket.Message.Send(ws, append([]byte{9}, meta...))
	if err != nil {
		log.Printf("Message.Send [stream][%s], [channel][%s] | [error]: %s\n", req.StreamID, req.ChannelID, err)
		return
	}

	err = websocket.Message.Send(ws, buffer)
	if err != nil {
		log.Printf("Message.Send [stream][%s], [channel][%s] | [error]: %s\n", req.StreamID, req.ChannelID, err)
		return
	}

	var videoStart bool
	var controlExit = make(chan bool, 10)
	go func() {
		defer func() {
			controlExit <- true
		}()
		for {
			var message string
			err := websocket.Message.Receive(ws, &message)
			if err != nil {
				log.Printf("Message.Receive [stream][%s], [channel][%s] | [error]: %s\n", req.StreamID, req.ChannelID, err)
				return
			}
		}
	}()

	noVideo := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-controlExit:
			log.Println("controlExit", "Client Reader Exit")

			return
		case <-noVideo.C:
			log.Println(ErrorStreamHasNoVideo())

			return
		case pck := <-subscriber.Packet:
			if pck.IsKeyFrame {
				noVideo.Reset(10 * time.Second)
				videoStart = true
			}

			if !videoStart {
				continue
			}

			ready, buf, err := muxerMSE.WritePacket(*pck, false)
			if err != nil {
				log.Printf("Message.Receive [stream][%s], [channel][%s] | [error]: %s\n", req.StreamID, req.ChannelID, err)
				return
			}

			if ready {
				err := ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err != nil {
					log.Printf("SetWriteDeadline [stream][%s], [channel][%s] | [error]: %s\n", req.StreamID, req.ChannelID, err)
					return
				}

				err = websocket.Message.Send(ws, buf)
				if err != nil {
					log.Printf("Message.Send [stream][%s], [channel][%s] | [error]: %s\n", req.StreamID, req.ChannelID, err)
					return
				}
			}
		}
	}
}
