//go:build integration

package geojs

import (
	"context"
	"testing"
)

func TestLookupIP_Live(t *testing.T) {
	c := NewClient(DefaultConfig())
	// 8.8.8.8 is Google DNS; always routable.
	info, err := c.LookupIP(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "8.8.8.8" {
		t.Errorf("IP = %q, want 8.8.8.8", info.IP)
	}
	if info.Country == "" {
		t.Error("Country is empty")
	}
	t.Logf("geo: %+v", info)
}

func TestMyGeo_Live(t *testing.T) {
	c := NewClient(DefaultConfig())
	info, err := c.MyGeo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.IP == "" {
		t.Error("IP is empty")
	}
	t.Logf("my geo: %+v", info)
}

func TestMyIP_Live(t *testing.T) {
	c := NewClient(DefaultConfig())
	info, err := c.MyIP(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.IP == "" {
		t.Error("IP is empty")
	}
	t.Logf("my IP: %s", info.IP)
}
