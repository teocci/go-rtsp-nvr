package webauth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type AuthRequest struct {
	Proto   string `json:"proto,omitempty"`
	Stream  string `json:"stream,omitempty"`
	Channel string `json:"channel,omitempty"`
	Token   string `json:"token,omitempty"`
	IP      string `json:"ip,omitempty"`
}

type AuthResponse struct {
	Status string `json:"status,omitempty"`
}

func RemoteAuthorization(req AuthRequest, token string) (ok bool) {
	buf, err := json.Marshal(&req)
	if err != nil {
		return
	}

	request, err := http.NewRequest("POST", token, bytes.NewBuffer(buf))
	if err != nil {
		return
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	var res AuthResponse
	err = json.Unmarshal(bodyBytes, &res)
	if err != nil {
		return
	}

	if res.Status == "1" {
		return true
	}

	return
}
