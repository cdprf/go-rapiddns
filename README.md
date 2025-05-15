# go-rapiddns

A cross-platform command-line tool to fetch and export DNS records from [rapiddns.io](https://rapiddns.io/) by IP address or domain. Supports CSV, TSV, JSON, and plain text output formats. Designed for automation, scripting, and bulk DNS data analysis.

## Features
- Query by IP address or domain
- Output in CSV (default), TSV, JSON, or TEXT format
- Verbose/debug mode for troubleshooting
- Accepts input via command-line argument or standard input (stdin)
- Cross-platform: Windows, Linux, macOS
- GitHub Actions workflow for multi-platform builds
- **Easy install via Go:**

## Installation

You can install directly from GitHub if you have Go 1.18+:

```sh
go install github.com/cdprf/go-rapiddns@latest
```

This will place the `go-rapiddns` binary in your `$GOPATH/bin` or `$GOBIN`.

## Usage

### Basic usage
```sh
go-rapiddns <ip-or-domain> [--format=csv|tsv|json|text] [--verbose]
```

### With stdin
You can pipe input directly:
```sh
echo "1.2.3.4" | go-rapiddns
cat iplist.txt | go-rapiddns
```
If no positional argument is provided, the tool reads the first non-empty line from stdin as the IP/domain.

### Output format
- `--format=csv`   (default)  Comma-separated values
- `--format=tsv`   Tab-separated values
- `--format=json`  JSON array
- `--format=text`  Space-separated values

### Verbose/debug mode
- `--verbose`   Prints each processed page and errors to the screen

### Example
```sh
go-rapiddns 8.8.8.8 --format=json --verbose
```

## Build

### Requirements
- Go 1.18+

### Build for your platform
```sh
go build -o go-rapiddns main.go
```

### Multi-platform build (via GitHub Actions)
On each push, binaries for Linux, Windows, and macOS will be built automatically. See `.github/workflows/build.yml`.

## License
MIT
