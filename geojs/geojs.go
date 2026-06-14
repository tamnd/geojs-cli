// Package geojs is the library behind the geojs command line:
// the HTTP client, request shaping, and typed data models for the GeoJS API
// (https://get.geojs.io/).
//
// No API key is required. The Client paces requests, sets a real User-Agent,
// and retries transient failures (429 and 5xx).
package geojs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Host is the API hostname.
const Host = "get.geojs.io"

// BaseURL is the root every API request is built from.
const BaseURL = "https://" + Host

// GeoInfo holds geolocation data for an IP address.
type GeoInfo struct {
	IP           string `json:"ip"           kit:"id"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
	Region       string `json:"region"`
	City         string `json:"city"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	Timezone     string `json:"timezone"`
	ASN          string `json:"asn"`
	Organization string `json:"organization"`
}

// IPInfo holds just the public IP address.
type IPInfo struct {
	IP string `json:"ip" kit:"id"`
}

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   BaseURL,
		UserAgent: "geojs-cli/0.1 (tamnd87@gmail.com)",
		Rate:      200 * time.Millisecond,
		Timeout:   15 * time.Second,
		Retries:   3,
	}
}

// Client talks to the GeoJS API.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client with the given configuration.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// LookupIP fetches geolocation info for a specific IP address.
func (c *Client) LookupIP(ctx context.Context, ip string) (*GeoInfo, error) {
	rawURL := fmt.Sprintf("%s/v1/ip/geo/%s.json", c.cfg.BaseURL, ip)
	return c.fetchGeo(ctx, rawURL)
}

// MyGeo fetches geolocation info for the caller's own IP address.
func (c *Client) MyGeo(ctx context.Context) (*GeoInfo, error) {
	rawURL := c.cfg.BaseURL + "/v1/ip/geo.json"
	return c.fetchGeo(ctx, rawURL)
}

// MyIP fetches only the caller's public IP address.
func (c *Client) MyIP(ctx context.Context) (*IPInfo, error) {
	rawURL := c.cfg.BaseURL + "/v1/ip.json"
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var info IPInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse ip: %w", err)
	}
	return &info, nil
}

// geoWire is the raw JSON shape returned by /v1/ip/geo endpoints. The city
// field can be null, so we decode it into a *string and then coerce to "".
type geoWire struct {
	IP           string  `json:"ip"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	Region       string  `json:"region"`
	City         *string `json:"city"`
	Latitude     string  `json:"latitude"`
	Longitude    string  `json:"longitude"`
	Timezone     string  `json:"timezone"`
	ASN          string  `json:"asn"`
	Organization string  `json:"organization"`
}

func (c *Client) fetchGeo(ctx context.Context, rawURL string) (*GeoInfo, error) {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var w geoWire
	if err := json.Unmarshal(body, &w); err != nil {
		return nil, fmt.Errorf("parse geo: %w", err)
	}
	city := ""
	if w.City != nil {
		city = *w.City
	}
	return &GeoInfo{
		IP:           w.IP,
		Country:      w.Country,
		CountryCode:  w.CountryCode,
		Region:       w.Region,
		City:         city,
		Latitude:     w.Latitude,
		Longitude:    w.Longitude,
		Timezone:     w.Timezone,
		ASN:          w.ASN,
		Organization: w.Organization,
	}, nil
}

// --- internal helpers ---

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
