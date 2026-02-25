package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
	"wabigo/config"
)

type Client struct {
	HTTP    *http.Client
	BaseURL string
	Secret  string
}

var (
	once sync.Once
	cfg  = config.Load()
)

func CreateHttpClient(baseURL string, secret string) *Client {
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
		Secret:  secret,
	}

}

func (c *Client) DoPostRequest(ctx context.Context, endpoint string, secret string, payload []byte) (*http.Response, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s%s", c.BaseURL, endpoint),
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error: %s", resp.Status)
	}
	//resp.Body.Close()
	return resp, nil
}

func (c *Client) DoGetRequest(ctx context.Context, endpoint string, secret string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s%s", c.BaseURL, endpoint),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	//resp.Body.Close()
	return resp, nil
}
