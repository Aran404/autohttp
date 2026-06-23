package http

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

// Client defines the behavior for sending HTTP requests, making it easily mockable.
type Client interface {
	Do(ctx context.Context, method, targetURL string, headers map[string]string, body io.Reader) (int, http.Header, []byte, error)
}

// Config configures the runtime HTTP client options.
type Config struct {
	ProxyURL    string
	Timeout     time.Duration
	InsecureTLS bool
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Ok         bool
}

// StdClient is a reusable, concurrent-safe HTTP client.
type StdClient struct {
	httpClient *http.Client
}

// NewStd creates and initializes a new StdClient with safe production defaults.
func NewStd(cfg Config) (*StdClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookiejar: %w", err)
	}

	// Clone the DefaultTransport to preserve vital connection pooling (Keep-Alive, MaxIdleConns)
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy url: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &StdClient{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
			Jar:       jar,
		},
	}, nil
}

// Do executes an HTTP request, manages the lifecycle of the response body, and returns the results.
func (c *StdClient) Do(ctx context.Context, method, targetURL string, headers map[string]string, body io.Reader) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       respBody,
		Ok:         resp.StatusCode >= 200 && resp.StatusCode <= 304,
	}, nil
}

// UnmarshalJSON is a generic helper to cleanly decode responses into concrete types.
func UnmarshalJSON[T any](body []byte) (T, error) {
	var target T
	if err := json.Unmarshal(body, &target); err != nil {
		return target, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return target, nil
}

// ExtractJSONPath parses a JSON response and extracts a raw string value using dot notation.
func ExtractJSONPath(body []byte, path string) (string, error) {
	var current any
	if err := json.Unmarshal(body, &current); err != nil {
		return "", fmt.Errorf("json parse error: %w", err)
	}

	parts := strings.Split(path, ".")
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("path %q broken: element at %q is not a JSON object", path, part)
		}

		val, ok := m[part]
		if !ok {
			return "", fmt.Errorf("path %q broken: key %q not found", path, part)
		}
		current = val
	}

	switch v := current.(type) {
	case string:
		return v, nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to remarshal slice/object path value: %w", err)
		}
		return string(b), nil
	}
}
