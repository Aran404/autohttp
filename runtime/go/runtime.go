package gort

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Config configures the runtime HTTP client.
type Config struct {
	ProxyURL    string
	Timeout     time.Duration
	InsecureTLS bool
}

// Client is a reusable HTTP client with cookie jar and state management.
type Client struct {
	httpClient *http.Client
	jar        *cookiejar.Jar
}

// New creates a new runtime Client.
func New(cfg Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar: %w", err)
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.InsecureTLS},
	}
	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("proxy url: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
			Jar:       jar,
		},
		jar: jar,
	}, nil
}

// Do sends an HTTP request and returns the response status, headers, and body.
func (c *Client) Do(ctx context.Context, method, rawURL string, headers map[string]string, body io.Reader) (int, map[string]string, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("new request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("read body: %w", err)
	}
	outHeaders := make(map[string]string)
	for k := range resp.Header {
		outHeaders[k] = resp.Header.Get(k)
	}
	return resp.StatusCode, outHeaders, respBody, nil
}

// ExtractJSON parses a JSON response and extracts a value at the given path.
// Path uses dot notation: "data.token" or "user.id".
func ExtractJSON(body []byte, path string) (string, error) {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("json parse: %w", err)
	}
	parts := strings.Split(path, ".")
	current := data
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("path %q: not an object at %q", path, part)
		}
		val, ok := m[part]
		if !ok {
			return "", fmt.Errorf("path %q: key %q not found", path, part)
		}
		current = val
	}
	switch v := current.(type) {
	case string:
		return v, nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("marshal value: %w", err)
		}
		return string(b), nil
	}
}
