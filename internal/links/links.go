package links

import (
	"fmt"
	"net/url"
	"strings"
)

var AllowedSignalSourceTypes = map[string]struct{}{
	"feedback":   {},
	"github":     {},
	"google_doc": {},
	"jira":       {},
	"linear":     {},
	"note":       {},
	"notion":     {},
	"prototype":  {},
	"release":    {},
	"slack":      {},
}

var AllowedArtifactLinkTypes = map[string]struct{}{
	"confluence":  {},
	"figma":       {},
	"github_issue": {},
	"github_pr":   {},
	"google_doc":  {},
	"generic":     {},
	"notion":      {},
	"slack_thread": {},
}

func NormalizeSignalSourceType(raw string) (string, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if _, ok := AllowedSignalSourceTypes[value]; !ok {
		return "", fmt.Errorf("invalid signal source type")
	}
	return value, nil
}

func NormalizeArtifactLinkType(raw string) (string, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return "", nil
	}
	if _, ok := AllowedArtifactLinkTypes[value]; !ok {
		return "", fmt.Errorf("invalid artifact link type")
	}
	return value, nil
}

func NormalizeExternalURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("invalid external url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid external url")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid external url")
	}
	return parsed.String(), nil
}

func ValidateTypedExternalLink(linkType string, rawURL string) (string, string, error) {
	normalizedType, err := NormalizeArtifactLinkType(linkType)
	if err != nil {
		return "", "", err
	}
	normalizedURL, err := NormalizeExternalURL(rawURL)
	if err != nil {
		return "", "", err
	}

	if (normalizedType == "") != (normalizedURL == "") {
		return "", "", fmt.Errorf("typed external links require both a type and a url")
	}

	return normalizedType, normalizedURL, nil
}
