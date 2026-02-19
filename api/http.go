package api

import (
	"net"
	"net/http"
	"time"
)

type Client struct {
	HTTP *http.Client
	BaseURL string
}

func NewHttpClient(baseURL string) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &Client{
		HTTP: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
		BaseURL: baseURL,
	}
}
