// Package session
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-19
package session

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func LoadData() (session *Session) {
	data, err := ioutil.ReadFile("config.json")
	if err == nil {
		err = json.Unmarshal(data, &session)
		if err != nil {
			log.Fatalln(err)
		}
		for i, s := range session.Streams {
			for j, ch := range s.Channels {
				ch.Subscribers = make(map[string]*Subscriber)
				session.Streams[i].Channels[j] = ch
			}
		}
	}

	return
}
