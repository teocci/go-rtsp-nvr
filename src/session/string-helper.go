// Package session
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-05
package session

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

// generateUUID makes random uuid for streams and subscribers
func generateUUID() (uuid string) {
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		log.Println("Error: ", err)

		return
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// StringToInt converts string into int if an error occurs returns zero
func StringToInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return n
}

// StringInBetween fin char to char sub string
func StringInBetween(str, start, end string) (r string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}

	str = str[s+len(start):]
	e := strings.Index(str, end)
	if e == -1 {
		return
	}

	return str[:e]
}
