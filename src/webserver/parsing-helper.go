// Package webserver
// Created by RTT.
// Author: teocci@yandex.com on 2021-Nov-03
package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

func extractJSONData(r io.ReadCloser, data *StreamRequest) {
	jsonString, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = json.Unmarshal(jsonString, &data)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// stringToInt convert string to int if err to zero
func stringToInt(val string) (i int) {
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	return
}

// stringInBetween fin char to char sub string
func stringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	str = str[s+len(start):]
	e := strings.Index(str, end)
	if e == -1 {
		return
	}
	str = str[:e]

	return str
}
