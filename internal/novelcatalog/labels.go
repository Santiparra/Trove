package novelcatalog

import (
	"regexp"
	"strings"
)

var volumeLabel = regexp.MustCompile(`(?i)\b(?:vol(?:ume)?\.?|tomo)\s*(\d+(?:[.,]\d+)?)`)
var japaneseVolumeLabel = regexp.MustCompile(`第?\s*(\d+(?:[.,]\d+)?)\s*巻`)
var partLabel = regexp.MustCompile(`(?i)\b(?:part|parte)\s*(\d+(?:\.\d+)?)`)
var yearLabel = regexp.MustCompile(`(?i)\b(?:year\s*(\d+)|(\d+)(?:st|nd|rd|th)\s+year)`)

// PrintedLabel is presentation metadata, never a reading count or a primary key.
func PrintedLabel(title, original string) string {
	label := ""
	if match := volumeLabel.FindStringSubmatch(title); len(match) > 1 {
		label = strings.ReplaceAll(match[1], ",", ".")
	} else if match := japaneseVolumeLabel.FindStringSubmatch(original); len(match) > 1 {
		label = strings.ReplaceAll(match[1], ",", ".")
	}
	if label == "" {
		return ""
	}
	if part := partLabel.FindStringSubmatch(title); len(part) > 1 {
		return "Parte " + part[1] + " · " + label
	}
	if year := yearLabel.FindStringSubmatch(title); len(year) > 2 {
		value := year[1]
		if value == "" {
			value = year[2]
		}
		return "Año " + value + " · " + label
	}
	return label
}
