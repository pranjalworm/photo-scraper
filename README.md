# photo-scraper

A CLI tool that downloads images from [Unsplash](https://unsplash.com) via their public API, preserving photographer attribution and copyright metadata alongside each image.

## What it does

`photo-scraper` lets you bulk-download Unsplash photos by search query. For every photo it downloads, it writes a companion `.json` sidecar file containing:

- Photographer name, username, and profile URL
- Photo description and tags
- EXIF data (camera make/model, aperture, shutter speed, focal length, ISO)
- Location (city, country if available)
- Source URL and license
- Download timestamp

This makes it suitable for building offline datasets, design assets, or any workflow where you need traceable attribution alongside the image files.

## How it works

1. Sends a search query to `api.unsplash.com/search/photos` (paginating up to `maxPerPage=30` per request).
2. For each result, downloads the image at the requested quality tier.
3. Writes a `photo-{id}.jpg` (or `.png`/`.webp`) and a `photo-{id}.json` sidecar to the output directory.
4. Fires a download-tracking request to Unsplash's `download_location` endpoint, as required by [Unsplash API guidelines](https://help.unsplash.com/en/articles/2511245-unsplash-api-guidelines).

## Prerequisites

- **Go 1.22+**
- An **Unsplash API access key** — register a free application at [unsplash.com/developers](https://unsplash.com/developers) to get one.

## Installation

```bash
git clone https://github.com/pranjaldubey/photo-scraper
cd photo-scraper
go build -o photo-scraper .
```

Or install directly:

```bash
go install github.com/pranjaldubey/photo-scraper@latest
```

## Configuration

The API key can be provided in three ways (in order of precedence):

1. `--api-key` flag on the command line
2. `UNSPLASH_ACCESS_KEY` environment variable
3. A `.env` file in the working directory containing `UNSPLASH_ACCESS_KEY=your_key_here`

`.env` is loaded automatically — no extra setup needed.

## Usage

```
photo-scraper download [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--query` | `-q` | *(required)* | Search query |
| `--count` | `-c` | `10` | Number of photos to download |
| `--output` | `-o` | `./photos` | Output directory |
| `--quality` | | `regular` | Image quality: `raw`, `full`, `regular`, `small`, `thumb` |
| `--api-key` | | | Unsplash API access key |
| `--dry-run` | | `false` | Preview results without downloading |

### Examples

```bash
# Download 10 mountain photos at regular quality (default)
photo-scraper download -q "mountains"

# Download 20 photos into a named folder
photo-scraper download -q "mountains" -c 20 -o ./mountains

# Download full-resolution street photography
photo-scraper download -q "street photography" --quality full

# Preview what would be downloaded without saving anything
photo-scraper download -q "nature" --dry-run

# Pass the API key inline
photo-scraper download -q "ocean" --api-key YOUR_KEY

# Or via env var
UNSPLASH_ACCESS_KEY=YOUR_KEY photo-scraper download -q "ocean"
```

## Output structure

```
photos/
├── photo-abc123.jpg
├── photo-abc123.json
├── photo-def456.jpg
├── photo-def456.json
└── ...
```

Example sidecar JSON:

```json
{
  "id": "abc123",
  "description": "brown rocky mountain under blue sky",
  "photographer": {
    "name": "Jane Doe",
    "username": "janedoe",
    "profile_url": "https://unsplash.com/@janedoe"
  },
  "source_url": "https://unsplash.com/photos/abc123",
  "exif": {
    "make": "Canon",
    "model": "EOS R5",
    "exposure_time": "1/500",
    "aperture": "5.6",
    "focal_length": "85.0",
    "iso": 400
  },
  "location": {
    "city": "Innsbruck",
    "country": "Austria"
  },
  "tags": ["mountain", "sky", "nature", "landscape"],
  "license": "Unsplash License (https://unsplash.com/license)",
  "downloaded_at": "2026-04-14T10:30:00Z"
}
```

## Development

```bash
# Run all tests
go test ./...

# Static analysis
go vet ./...

# Tidy dependencies
go mod tidy
```

## License

Images downloaded through this tool are subject to the [Unsplash License](https://unsplash.com/license). Always credit the photographer when using images.
