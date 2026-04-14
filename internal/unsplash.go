package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	unsplashAPIBase = "https://api.unsplash.com"
	unsplashWebBase = "https://unsplash.com"
	unsplashLicense = "Unsplash License (https://unsplash.com/license)"
	maxPerPage      = 30
)

// UnsplashClient interfaces with the Unsplash API.
type UnsplashClient struct {
	http      *http.Client
	accessKey string
}

func NewUnsplashClient(accessKey string) *UnsplashClient {
	return &UnsplashClient{
		http:      &http.Client{Timeout: 30 * time.Second},
		accessKey: accessKey,
	}
}

// unsplash API response types

type unsplashSearchResponse struct {
	Total      int             `json:"total"`
	TotalPages int             `json:"total_pages"`
	Results    []unsplashPhoto `json:"results"`
}

type unsplashPhoto struct {
	ID             string        `json:"id"`
	Description    string        `json:"description"`
	AltDescription string        `json:"alt_description"`
	URLs           unsplashURLs  `json:"urls"`
	Links          unsplashLinks `json:"links"`
	User           unsplashUser  `json:"user"`
	EXIF           *unsplashEXIF `json:"exif"`
	Location       *unsplashLoc  `json:"location"`
	Tags           []unsplashTag `json:"tags"`
}

type unsplashURLs struct {
	Raw     string `json:"raw"`
	Full    string `json:"full"`
	Regular string `json:"regular"`
	Small   string `json:"small"`
	Thumb   string `json:"thumb"`
}

type unsplashLinks struct {
	Self             string `json:"self"`
	HTML             string `json:"html"`
	Download         string `json:"download"`
	DownloadLocation string `json:"download_location"`
}

type unsplashUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Bio      string `json:"bio"`
	Links    struct {
		HTML string `json:"html"`
	} `json:"links"`
}

type unsplashEXIF struct {
	Make         string `json:"make"`
	Model        string `json:"model"`
	ExposureTime string `json:"exposure_time"`
	Aperture     string `json:"aperture"`
	FocalLength  string `json:"focal_length"`
	ISO          int    `json:"iso"`
}

type unsplashLoc struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

type unsplashTag struct {
	Title string `json:"title"`
}

// DownloadConfig holds parameters for a download session.
type DownloadConfig struct {
	Query   string
	Count   int
	Output  string
	Quality string // raw, full, regular, small, thumb
	DryRun  bool
}

// Download searches for photos and downloads them with metadata.
func (uc *UnsplashClient) Download(ctx context.Context, cfg DownloadConfig) error {
	fmt.Printf("Searching Unsplash for %q (requesting %d photos)...\n", cfg.Query, cfg.Count)

	photos, err := uc.search(ctx, cfg.Query, cfg.Count)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(photos) == 0 {
		fmt.Println("No photos found.")
		return nil
	}

	fmt.Printf("Found %d photos. ", len(photos))
	if cfg.DryRun {
		fmt.Println("Dry run — listing without downloading:")
		for i, p := range photos {
			desc := p.AltDescription
			if desc == "" {
				desc = p.Description
			}
			fmt.Printf("  %d. [%s] %q by %s\n", i+1, p.ID, desc, p.User.Name)
		}
		return nil
	}
	fmt.Println("Downloading...")

	if err := os.MkdirAll(cfg.Output, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	var downloaded int
	for i, photo := range photos {
		fmt.Printf("  [%d/%d] %s by %s... ", i+1, len(photos), photo.ID, photo.User.Name)

		if err := uc.downloadPhoto(ctx, photo, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "FAILED: %v\n", err)
			continue
		}
		downloaded++
		fmt.Println("OK")

		// Trigger download tracking as required by Unsplash API guidelines.
		uc.triggerDownload(ctx, photo)
	}

	fmt.Printf("\nDone. Downloaded %d/%d photos to %s\n", downloaded, len(photos), cfg.Output)
	return nil
}

func (uc *UnsplashClient) search(ctx context.Context, query string, count int) ([]unsplashPhoto, error) {
	var allPhotos []unsplashPhoto
	remaining := count
	page := 1

	for remaining > 0 {
		perPage := min(remaining, maxPerPage)

		params := url.Values{
			"query":    {query},
			"page":     {strconv.Itoa(page)},
			"per_page": {strconv.Itoa(perPage)},
		}

		endpoint := unsplashAPIBase + "/search/photos?" + params.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		uc.setAuthHeader(req)

		resp, err := uc.http.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
		}

		var searchResp unsplashSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		resp.Body.Close()

		if len(searchResp.Results) == 0 {
			break
		}

		allPhotos = append(allPhotos, searchResp.Results...)
		remaining -= len(searchResp.Results)
		page++

		if page > searchResp.TotalPages {
			break
		}
	}

	if len(allPhotos) > count {
		allPhotos = allPhotos[:count]
	}
	return allPhotos, nil
}

func (uc *UnsplashClient) downloadPhoto(ctx context.Context, photo unsplashPhoto, cfg DownloadConfig) error {
	imageURL := uc.pickQuality(photo.URLs, cfg.Quality)
	if imageURL == "" {
		return fmt.Errorf("no URL for quality %q", cfg.Quality)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return err
	}

	resp, err := uc.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("image download returned %d", resp.StatusCode)
	}

	ext := extensionFromContentType(resp.Header.Get("Content-Type"))
	imagePath := filepath.Join(cfg.Output, fmt.Sprintf("photo-%s%s", photo.ID, ext))

	f, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(imagePath)
		return err
	}
	f.Close()

	meta := uc.buildMetadata(photo)
	metaPath := filepath.Join(cfg.Output, fmt.Sprintf("photo-%s.json", photo.ID))
	metaJSON, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}
	return os.WriteFile(metaPath, metaJSON, 0o644)
}

func (uc *UnsplashClient) buildMetadata(photo unsplashPhoto) Photo {
	desc := photo.AltDescription
	if desc == "" {
		desc = photo.Description
	}

	meta := Photo{
		ID:          photo.ID,
		Description: desc,
		Photographer: Photographer{
			Name:       photo.User.Name,
			Username:   photo.User.Username,
			ProfileURL: photo.User.Links.HTML,
			Bio:        photo.User.Bio,
		},
		SourceURL:    unsplashWebBase + "/photos/" + photo.ID,
		License:      unsplashLicense,
		DownloadedAt: time.Now().UTC(),
	}

	if photo.EXIF != nil {
		meta.EXIF = &EXIF{
			Make:         photo.EXIF.Make,
			Model:        photo.EXIF.Model,
			ExposureTime: photo.EXIF.ExposureTime,
			Aperture:     photo.EXIF.Aperture,
			FocalLength:  photo.EXIF.FocalLength,
			ISO:          photo.EXIF.ISO,
		}
	}

	if photo.Location != nil && (photo.Location.City != "" || photo.Location.Country != "") {
		meta.Location = &Location{
			City:    photo.Location.City,
			Country: photo.Location.Country,
		}
	}

	for _, tag := range photo.Tags {
		if tag.Title != "" {
			meta.Tags = append(meta.Tags, tag.Title)
		}
	}

	return meta
}

func (uc *UnsplashClient) triggerDownload(ctx context.Context, photo unsplashPhoto) {
	if photo.Links.DownloadLocation == "" {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, photo.Links.DownloadLocation, nil)
	if err != nil {
		return
	}
	uc.setAuthHeader(req)
	resp, err := uc.http.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (uc *UnsplashClient) setAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Client-ID "+uc.accessKey)
}

func (uc *UnsplashClient) pickQuality(urls unsplashURLs, quality string) string {
	switch strings.ToLower(quality) {
	case "raw":
		return urls.Raw
	case "full":
		return urls.Full
	case "regular":
		return urls.Regular
	case "small":
		return urls.Small
	case "thumb":
		return urls.Thumb
	default:
		return urls.Regular
	}
}

func extensionFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "webp"):
		return ".webp"
	case strings.Contains(ct, "gif"):
		return ".gif"
	default:
		return ".jpg"
	}
}
