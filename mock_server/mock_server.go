package server

import (
	"bytes"
	// "encoding/json"
	"fmt"
	"io"
	// "io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

var mockSlack *MockSlack

type MockSlack struct {
	Server *httptest.Server
}

func New() *MockSlack {
	mockSlack = &MockSlack{Server: mockServer()}
	return mockSlack
}

func mockServer() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/chat.postMessage", handlePostMessage)

	return httptest.NewServer(handler)
}

func IOCopy(reader io.Reader) ([]byte, error) {
	var (
		buf  bytes.Buffer
		_, _ = io.Copy(&buf, reader)
	)
	return buf.Bytes(), nil
}

func handlePostMessage(w http.ResponseWriter, r *http.Request) {
	body, _ := IOCopy(r.Body)
	kvs := strings.Split(string(body), "&")
	log.Printf("Body stringified %v", string(body))

	log.Printf("KVS splitted %v", kvs)

	if len(kvs) < 2 {
		log.Printf("KVS length %v", len(kvs))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m := make(map[string]string)

	for _, s := range kvs {
		kv := strings.Split(s, "=")
		s, err := url.QueryUnescape(kv[1])
		if err != nil {
			m[kv[0]] = kv[1]
		} else {
			m[kv[0]] = s
		}
	}

	log.Printf("KVS decoded %v", m)

	// ref: https://api.slack.com/methods/chat.postMessage
	const response = `{
    "ok": true,
    "channel": "%s",
    "ts": "0000",
    "message": {
        "text": "%s",
        "username": "ecto1",
        "bot_id": "B19LU7CSY",
        "attachments": [
            {
                "text": "This is an attachment",
                "id": 1,
                "fallback": "This is an attachment's fallback"
            }
        ],
        "type": "message",
        "subtype": "bot_message",
        "ts": "1503435956.000247"
    }
 }`

	s := fmt.Sprintf(response, m["channel"], m["text"])
	_, _ = w.Write([]byte(s))
}
