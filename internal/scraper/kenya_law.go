package scraper

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly/v2"
)

type ScrapedSegment struct {
	RoadName      string  `json:"road_name"`
	SpeedLimitKmh int     `json:"speed_limit_kmh"`
	RoadClass     string  `json:"road_class"`
	Source        string  `json:"source"`
	County        string  `json:"county"`
	Description   string  `json:"description"`
	StartPoint    string  `json:"start_point"`
	EndPoint      string  `json:"end_point"`
}

func ScrapeKenyaLaw() ([]ScrapedSegment, error) {
	var segments []ScrapedSegment

	c := colly.NewCollector(
		colly.AllowedDomains("www.kenyalaw.org", "kenyalaw.org"),
	)

	c.OnHTML("div.akn-section", func(e *colly.HTMLElement) {
		text := e.Text
		if containsSpeedReference(text) {
			seg := parseSpeedSection(text)
			if seg != nil {
				segments = append(segments, *seg)
			}
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Scrape error on %s: %v", r.Request.URL, err)
	})

	// Traffic Act Cap 403
	err := c.Visit("https://www.kenyalaw.org/kl/fileadmin/pdfdownloads/Acts/TrafficAct_Cap403.pdf")
	if err != nil {
		return segments, fmt.Errorf("visiting Kenya Law: %w", err)
	}

	segments = append(segments, defaultSpeedLimits()...)

	return segments, nil
}

func containsSpeedReference(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "speed limit") ||
		strings.Contains(lower, "kilometres per hour") ||
		strings.Contains(lower, "km/h")
}

func parseSpeedSection(text string) *ScrapedSegment {
	// Placeholder for more sophisticated parsing.
	// Real implementation would use regex to extract road names,
	// speed values, and location references from gazette text.
	return nil
}

func defaultSpeedLimits() []ScrapedSegment {
	// Kenya Traffic Act Cap 403 — default speed limits by vehicle/road type.
	return []ScrapedSegment{
		{
			RoadName:      "Default — Built-up areas",
			SpeedLimitKmh: 50,
			RoadClass:     "urban",
			Source:        "Traffic Act Cap 403, Section 42",
			Description:   "Default speed limit in built-up areas for all motor vehicles",
		},
		{
			RoadName:      "Default — Non built-up areas (light vehicles)",
			SpeedLimitKmh: 110,
			RoadClass:     "highway",
			Source:        "Traffic Act Cap 403, Section 42",
			Description:   "Default speed limit outside built-up areas for saloon cars, pick-ups, land rovers",
		},
		{
			RoadName:      "Default — Non built-up areas (commercial vehicles)",
			SpeedLimitKmh: 80,
			RoadClass:     "highway",
			Source:        "Traffic Act Cap 403, Section 42",
			Description:   "Default speed limit outside built-up areas for commercial vehicles, matatus",
		},
	}
}
