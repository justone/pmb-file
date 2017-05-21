package main

import "net/http"

// UrlAvailableMessage is a convenient struct to deserialize upload and
// download url messages
type UrlAvailableMessage struct {
	Type      string      `json:"type"`
	Requestor string      `json:"requestor"`
	Filename  string      `json:"filename"`
	Url       string      `json:"url"`
	Header    http.Header `json:"headers"`
}
