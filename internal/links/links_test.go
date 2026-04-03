package links

import "testing"

func TestNormalizeSignalSourceType(t *testing.T) {
	t.Parallel()

	value, err := NormalizeSignalSourceType(" GitHub ")
	if err != nil {
		t.Fatalf("expected valid source type: %v", err)
	}
	if value != "github" {
		t.Fatalf("expected github, got %q", value)
	}

	if _, err := NormalizeSignalSourceType("unsupported"); err == nil {
		t.Fatal("expected invalid signal source type")
	}
}

func TestValidateTypedExternalLink(t *testing.T) {
	t.Parallel()

	linkType, externalURL, err := ValidateTypedExternalLink("github_pr", "https://github.com/example/repo/pull/1")
	if err != nil {
		t.Fatalf("expected valid typed link: %v", err)
	}
	if linkType != "github_pr" || externalURL == "" {
		t.Fatalf("expected normalized typed link, got %q %q", linkType, externalURL)
	}

	if _, _, err := ValidateTypedExternalLink("github_pr", ""); err == nil {
		t.Fatal("expected missing url to fail")
	}
	if _, _, err := ValidateTypedExternalLink("", "https://github.com/example/repo/pull/1"); err == nil {
		t.Fatal("expected missing link type to fail")
	}
	if _, _, err := ValidateTypedExternalLink("github_pr", "mailto:test@example.com"); err == nil {
		t.Fatal("expected invalid url scheme to fail")
	}
}
