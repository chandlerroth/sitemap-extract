# Sitemap URL Extractor

A command-line tool written in Go to extract URLs from XML sitemaps. Supports:
- Standard XML sitemaps
- Sitemap index files
- Gzipped sitemaps
- Export to file or stdout

## Installation

```bash
# Clone the repository
git clone https://github.com/chandlerroth/sitemap-extract
cd sitemap-extractor

# Build the binary
go build -o sitemap
```

## Usage

Basic usage to print URLs to stdout:
```bash
./sitemap https://example.com/sitemap.xml
```

Save URLs to a file:
```bash
./sitemap -o urls.txt https://example.com/sitemap.xml
```

## Features

- Extracts URLs from standard XML sitemaps
- Supports sitemap index files (sitemapindex)
- Handles gzipped sitemaps automatically
- Can export URLs to a file or print to stdout
- Follows the [Sitemaps XML format](https://www.sitemaps.org/protocol.html) protocol

## Requirements

- Go 1.21 or later

## License

MIT License
