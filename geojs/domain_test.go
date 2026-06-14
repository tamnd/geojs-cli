package geojs

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

// These tests are offline: they exercise the URI driver's pure string
// functions and the host wiring, which need no network. Live HTTP behaviour
// is covered in geojs_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "geojs" {
		t.Errorf("Scheme = %q, want geojs", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "geojs" {
		t.Errorf("Identity.Binary = %q, want geojs", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		in, typ, id string
	}{
		{"8.8.8.8", "ip", "8.8.8.8"},
		{"1.1.1.1", "ip", "1.1.1.1"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("Classify(\"\") should return an error")
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("ip", "8.8.8.8")
	want := "https://get.geojs.io/v1/ip/geo/8.8.8.8.json"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("page", "foo")
	if err == nil {
		t.Error("Locate with unknown type should return an error")
	}
}

// TestHostWiring mounts the driver in a kit Host and checks that a GeoInfo
// record mints to its URI correctly. The URI type is the lowercased struct
// name ("geoinfo") as kit derives it automatically.
func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}

	g := &GeoInfo{
		IP:          "8.8.8.8",
		Country:     "United States",
		CountryCode: "US",
	}
	u, err := h.Mint(g)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	// kit derives the URI type from the struct name: GeoInfo -> geoinfo.
	if want := "geojs://geoinfo/8.8.8.8"; u.String() != want {
		t.Errorf("Mint = %q, want %q", u.String(), want)
	}
}
