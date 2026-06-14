package geojs

import (
	"context"
	"fmt"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// init registers the geojs driver so a multi-domain host (ant) can enable it
// with a single blank import:
//
//	import _ "github.com/tamnd/geojs-cli/geojs"
//
// The same Domain also builds the standalone geojs binary (see cli.NewApp),
// so binary and host share one source of truth.
func init() { kit.Register(Domain{}) }

// Domain is the geojs driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "geojs",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "geojs",
			Short:  "IP geolocation and public-IP lookup via get.geojs.io.",
			Long: `IP geolocation and public-IP lookup via get.geojs.io.

geojs reads public GeoJS data over HTTPS, shapes it into clean records, and
prints output that pipes into the rest of your tools. No API key, nothing to
run alongside it.`,
			Site: Host,
			Repo: "https://github.com/tamnd/geojs-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "lookup", Group: "read", Single: true,
		Summary: "Geo lookup for your IP or a specified IP",
		Args: []kit.Arg{
			{Name: "ip", Help: "IP address to look up (optional; omit for your own IP)", Optional: true},
		}}, lookupIP)

	kit.Handle(app, kit.OpMeta{Name: "myip", Group: "read", Single: true,
		Summary: "Get your public IP address"}, myIP)
}

// newClient builds the Client from kit config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type lookupInput struct {
	IP     string  `kit:"arg" help:"IP address to look up (optional; omit for your own IP)"`
	Client *Client `kit:"inject"`
}

type myIPInput struct {
	Client *Client `kit:"inject"`
}

// --- handlers ---

func lookupIP(ctx context.Context, in lookupInput, emit func(*GeoInfo) error) error {
	var (
		info *GeoInfo
		err  error
	)
	if in.IP == "" {
		info, err = in.Client.MyGeo(ctx)
	} else {
		info, err = in.Client.LookupIP(ctx, in.IP)
	}
	if err != nil {
		return err
	}
	return emit(info)
}

func myIP(ctx context.Context, in myIPInput, emit func(*IPInfo) error) error {
	info, err := in.Client.MyIP(ctx)
	if err != nil {
		return err
	}
	return emit(info)
}

// Classify turns an IP address string into (type, id).
func (Domain) Classify(input string) (string, string, error) {
	if input == "" {
		return "", "", errs.Usage("geojs: expected an IP address or resource reference")
	}
	return "ip", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(t, id string) (string, error) {
	switch t {
	case "ip":
		return fmt.Sprintf("https://%s/v1/ip/geo/%s.json", Host, id), nil
	default:
		return "", errs.Usage("geojs has no resource type %q", t)
	}
}
