# geojs

IP geolocation and public-IP lookup via [get.geojs.io](https://get.geojs.io/).

`geojs` is a single pure-Go binary. It reads public GeoJS data over plain
HTTPS, shapes it into clean records, and prints output that pipes into the rest
of your tools. No API key, nothing to run alongside it.

The same package is also a [resource-URI driver](#use-it-as-a-resource-uri-driver),
so a host program like [ant](https://github.com/tamnd/ant) can address
geojs as `geojs://` URIs.

## Install

```bash
go install github.com/tamnd/geojs-cli/cmd/geojs@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/geojs-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/geojs:latest --help
```

## Usage

```bash
geojs lookup              # geo info for your own IP
geojs lookup 8.8.8.8     # geo info for a specific IP
geojs myip               # just your public IP address
geojs lookup 8.8.8.8 -o json   # as JSON, ready for jq
geojs --help             # the whole command tree
```

Every command shares one output contract: `-o table|json|jsonl|csv|tsv|url|raw`,
`--fields` to pick columns, `--template` for a custom line, and `-n` to limit.
The default adapts to where output goes (a table on a terminal, JSONL in a
pipe), so the same command reads well by hand and parses cleanly downstream.

## Commands

| Command | Description |
|---|---|
| `geojs lookup [ip]` | Geo lookup for your IP or a specified IP |
| `geojs myip` | Get your public IP address |

## Serve it

The same operations are available over HTTP and as an MCP tool set for agents,
with no extra code:

```bash
geojs serve --addr :7777    # GET /v1/lookup  returns NDJSON
geojs mcp                   # speak MCP over stdio
```

## Use it as a resource-URI driver

`geojs` registers a `geojs` domain the way a program registers a
database driver with `database/sql`. A host enables it with one blank import:

```go
import _ "github.com/tamnd/geojs-cli/geojs"
```

Then [ant](https://github.com/tamnd/ant) (or any program that links the package)
dereferences `geojs://` URIs without knowing anything about geojs:

```bash
ant get geojs://geoinfo/8.8.8.8   # fetch the record
ant url geojs://geoinfo/8.8.8.8   # the live https URL
```

## Development

```
cmd/geojs/   thin main: hands cli.NewApp to kit.Run
cli/         assembles the kit App from the geojs domain
geojs/       the library: HTTP client, data models, and domain.go (the driver)
docs/        tago documentation site
```

```bash
make build      # ./bin/geojs
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
