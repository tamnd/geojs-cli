package geojs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLookupIP_Unit(t *testing.T) {
	payload := geoWire{
		IP:           "8.8.8.8",
		Country:      "United States",
		CountryCode:  "US",
		Region:       "California",
		Latitude:     "37.751",
		Longitude:    "-97.822",
		Timezone:     "America/Chicago",
		ASN:          "15169",
		Organization: "AS15169 GOOGLE",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	c := NewClient(cfg)

	info, err := c.LookupIP(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "8.8.8.8" {
		t.Errorf("IP = %q, want 8.8.8.8", info.IP)
	}
	if info.Country != "United States" {
		t.Errorf("Country = %q, want United States", info.Country)
	}
	if info.CountryCode != "US" {
		t.Errorf("CountryCode = %q, want US", info.CountryCode)
	}
	if info.Latitude != "37.751" {
		t.Errorf("Latitude = %q, want 37.751", info.Latitude)
	}
}

func TestLookupIP_NullCity(t *testing.T) {
	// city can be JSON null; GeoInfo.City should become "".
	raw := `{"ip":"1.1.1.1","country":"Australia","country_code":"AU","region":"","city":null,"latitude":"-33.494","longitude":"143.2104","timezone":"Australia/Sydney","asn":"13335","organization":"AS13335 CLOUDFLARENET"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(raw))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	c := NewClient(cfg)

	info, err := c.LookupIP(context.Background(), "1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	if info.City != "" {
		t.Errorf("City = %q, want empty string for null JSON", info.City)
	}
}

func TestMyIP_Unit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"203.0.113.42"}`))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	c := NewClient(cfg)

	info, err := c.MyIP(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "203.0.113.42" {
		t.Errorf("IP = %q, want 203.0.113.42", info.IP)
	}
}

func TestRetryOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"8.8.8.8","country":"United States","country_code":"US","region":"","city":null,"latitude":"37.751","longitude":"-97.822","timezone":"America/Chicago","asn":"15169","organization":"AS15169 GOOGLE"}`))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := NewClient(cfg)

	start := time.Now()
	info, err := c.LookupIP(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "8.8.8.8" {
		t.Errorf("IP = %q after retries", info.IP)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}
