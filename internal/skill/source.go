package skill

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var shorthandPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`)

// SourceRef holds a parsed skill source.
type SourceRef struct {
	Raw             string
	SourceType      string
	CanonicalSource string
	CloneURL        string
}

// ParseSource parses a source string into a SourceRef.
// Supports GitHub/GitLab shorthand, SCP notation, and HTTP(S)/SSH URLs.
func ParseSource(raw string) (SourceRef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return SourceRef{}, fmt.Errorf("parse source: empty source")
	}

	if shorthandPattern.MatchString(raw) {
		return SourceRef{
			Raw:             raw,
			SourceType:      "github",
			CanonicalSource: raw,
			CloneURL:        "https://github.com/" + raw + ".git",
		}, nil
	}

	if strings.HasPrefix(raw, "git@") {
		return parseSCPSource(raw)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return SourceRef{}, fmt.Errorf("parse source %q: %w", raw, err)
	}

	if u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ssh" {
		return parseURLSource(raw, u)
	}

	if u.Scheme == "file" {
		return SourceRef{
			Raw:             raw,
			SourceType:      "git",
			CanonicalSource: raw,
			CloneURL:        raw,
		}, nil
	}

	return SourceRef{}, fmt.Errorf("parse source %q: unsupported format", raw)
}

func parseSCPSource(raw string) (SourceRef, error) {
	rest, found := strings.CutPrefix(raw, "git@")
	if !found {
		return SourceRef{}, fmt.Errorf("parse source %q: invalid scp url", raw)
	}

	host, path, ok := strings.Cut(rest, ":")
	if !ok || host == "" || path == "" {
		return SourceRef{}, fmt.Errorf("parse source %q: invalid scp url", raw)
	}

	path = strings.TrimSuffix(path, ".git")

	if host == "github.com" || host == "gitlab.com" {
		return SourceRef{
			Raw:             raw,
			SourceType:      sourceTypeForHost(host),
			CanonicalSource: path,
			CloneURL:        raw,
		}, nil
	}

	return SourceRef{
		Raw:             raw,
		SourceType:      "git",
		CanonicalSource: raw,
		CloneURL:        raw,
	}, nil
}

func parseURLSource(raw string, u *url.URL) (SourceRef, error) {
	host := strings.ToLower(u.Hostname())
	path := strings.Trim(strings.TrimSuffix(u.EscapedPath(), ".git"), "/")
	if path == "" {
		return SourceRef{}, fmt.Errorf("parse source %q: missing repository path", raw)
	}

	if host == "github.com" || host == "gitlab.com" {
		return SourceRef{
			Raw:             raw,
			SourceType:      sourceTypeForHost(host),
			CanonicalSource: path,
			CloneURL:        "https://" + host + "/" + path + ".git",
		}, nil
	}

	return SourceRef{
		Raw:             raw,
		SourceType:      "git",
		CanonicalSource: raw,
		CloneURL:        raw,
	}, nil
}

func sourceTypeForHost(host string) string {
	if host == "gitlab.com" {
		return "gitlab"
	}

	return "github"
}

// SupportedSourceType reports whether the source type can be cloned/updated.
func SupportedSourceType(sourceType string) bool {
	switch sourceType {
	case "github", "gitlab", "git":
		return true
	default:
		return false
	}
}
