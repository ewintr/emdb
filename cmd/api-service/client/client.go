package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ewintr.nl/emdb/movie"
)

type Client struct {
	apiKey string
	url    string
	c      *http.Client
}

func New(url, apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		url:    url,
		c:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Store(m *movie.Movie) error {
	url := fmt.Sprintf("%s/movie/%s", c.url, m.ID)
	bodyJSON, err := json.Marshal(m)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(bodyJSON)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
