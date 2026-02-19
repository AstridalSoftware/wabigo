package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"wabigo/api"
)

type IssuesClient struct {
	*api.Client
}

type IssueCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageB64 string `json:"image_b64"`
}

type Issue struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Description string `json:"description"`
	OwnerID string `json:"owner_id"`
	CreatedAt string `json:"created_at"`
}

type IssuesResponse struct {
    Data  []Issue `json:"data"`
    Count int    `json:"count"`
}

func NewIssuesClient(baseURL string) *IssuesClient {
	return &IssuesClient{
		Client: api.NewHttpClient(baseURL),
	}
}

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

	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzIxMjU1NDgsInN1YiI6IjA5ZjIxYTczLWM4MTUtNDczMy04ZjI1LTU5ZDhmNjM1NmU1YyJ9.UyOQJRhjHu1NYXSrZbAiSFIyYk3-5-Gq1gzNkDXpzYQ")
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

func (c *IssuesClient) 	AddIssue(ctx context.Context, issue *IssueCreateRequest) (*Issue, error) {
	jsonBytes, err := json.Marshal(issue)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/issues", c.BaseURL),
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzIyMTIxNjUsInN1YiI6IjBjYWJhNjJlLTJkMGUtNGJhMC04MGE1LTE5YTcxZmFhMTM3NyJ9.MOzybLq-m3f9U9BLhDO2uMAFBSUc3RB3N_cgfjmDRc0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error: %s", resp.Status)
	}

	var response Issue
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
