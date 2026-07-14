// Package svcclient servislararo ichki HTTP chaqiruvlar uchun yupqa client.
// Har chaqiruvga X-Internal-Key qo'shadi va JSON javobni dst'ga ochadi.
package svcclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func New(baseURL, internalKey string) *Client {
	return &Client{
		baseURL:     baseURL,
		internalKey: internalKey,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

var ErrNotFound = fmt.Errorf("svcclient: resource not found")

// Get baseURL+path'ga so'rov yuborib, javob body'sini dst'ga unmarshal qiladi.
func (c *Client) Get(ctx context.Context, path string, dst any) error {
	return c.do(ctx, http.MethodGet, path, dst)
}

// Post body'siz POST yuboradi (masalan, hisoblagichni oshirish).
// dst nil bo'lishi mumkin.
func (c *Client) Post(ctx context.Context, path string, dst any) error {
	return c.do(ctx, http.MethodPost, path, dst)
}

func (c *Client) do(ctx context.Context, method, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-Internal-Key", c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("svcclient: %s %s returned %d", method, path, resp.StatusCode)
	}

	if dst == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}
