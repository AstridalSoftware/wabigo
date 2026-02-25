package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"wabigo/api"
	"wabigo/config"
)

type IssuesClient struct {
	*api.Client
}

type IssueCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageB64    string `json:"image_b64"`
}

type Issue struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	CreatedAt   string `json:"created_at"`
}

type IssuesResponse struct {
	Data  []Issue `json:"data"`
	Count int     `json:"count"`
}

func NewIssuesService(apiClient *api.Client) *IssuesClient {
	return &IssuesClient{
		Client: apiClient,
	}
}

var cfg = config.Load()

func (c *IssuesClient) GetAllIssues(ctx context.Context) (*IssuesResponse, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/issues", c.BaseURL),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", cfg.API_KEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error: %s", resp.Status)
	}

	var response IssuesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *IssuesClient) AddIssue(ctx context.Context, issue *IssueCreateRequest) (*Issue, error) {
	jsonBytes, err := json.Marshal(issue)
	if err != nil {
		return nil, err
	}
	fmt.Println("INICIO REQUEST")
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/issues", c.BaseURL),
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", cfg.API_KEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error: %s", resp.Status)
	}
	fmt.Println("FIN REQUEST")

	var response Issue
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	resp.Body.Close()
	fmt.Printf("Issue creado con ID: %s\n", response.ID)
	return &response, nil
}
