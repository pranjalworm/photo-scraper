package internal

import "time"

// Photo represents a downloaded photo with all its metadata.
type Photo struct {
	ID          string       `json:"id"`
	Description string       `json:"description,omitempty"`
	Photographer Photographer `json:"photographer"`
	SourceURL   string       `json:"source_url"`
	EXIF        *EXIF        `json:"exif,omitempty"`
	Location    *Location    `json:"location,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	License     string       `json:"license"`
	DownloadedAt time.Time   `json:"downloaded_at"`
}

// Photographer holds the credit/attribution information.
type Photographer struct {
	Name       string `json:"name"`
	Username   string `json:"username"`
	ProfileURL string `json:"profile_url"`
	Bio        string `json:"bio,omitempty"`
}

// EXIF holds camera/technical metadata.
type EXIF struct {
	Make         string `json:"make,omitempty"`
	Model        string `json:"model,omitempty"`
	ExposureTime string `json:"exposure_time,omitempty"`
	Aperture     string `json:"aperture,omitempty"`
	FocalLength  string `json:"focal_length,omitempty"`
	ISO          int    `json:"iso,omitempty"`
}

// Location holds geographic information.
type Location struct {
	City    string `json:"city,omitempty"`
	Country string `json:"country,omitempty"`
}
