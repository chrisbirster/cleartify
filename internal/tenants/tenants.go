package tenants

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type Mode string

const (
	ModeUnknown   Mode = "unknown"
	ModeMarketing Mode = "marketing"
	ModeApp       Mode = "app"
	ModeTenant    Mode = "tenant"
)

var (
	ErrInvalidHost   = errors.New("invalid host")
	ErrUnknownTenant = errors.New("unknown tenant")
)

type Organization struct {
	ID   string
	Slug string
	Name string
}

type Resolution struct {
	Mode         Mode
	Host         string
	Slug         string
	Organization *Organization
}

type LookupFunc func(ctx context.Context, slug string) (*Organization, error)

type Resolver struct {
	baseDomain    string
	appHost       string
	appHostname   string
	marketingHost string
	marketingName string
	wwwHost       string
}

type contextKey string

const tenantContextKey contextKey = "cleartify_tenant"

func NewResolver(baseDomain, appPublicURL, marketingURL string) (*Resolver, error) {
	baseDomain = normalizeHost(baseDomain)
	if baseDomain == "" {
		return nil, ErrInvalidHost
	}

	appURL, err := url.Parse(strings.TrimSpace(appPublicURL))
	if err != nil {
		return nil, err
	}

	marketingParsed, err := url.Parse(strings.TrimSpace(marketingURL))
	if err != nil {
		return nil, err
	}

	return &Resolver{
		baseDomain:    baseDomain,
		appHost:       normalizeAuthority(appURL.Host),
		appHostname:   normalizeHost(appURL.Host),
		marketingHost: normalizeAuthority(marketingParsed.Host),
		marketingName: normalizeHost(marketingParsed.Host),
		wwwHost:       "www." + baseDomain,
	}, nil
}

func (r *Resolver) ResolveHost(rawHost string) (Resolution, error) {
	authority := normalizeAuthority(rawHost)
	host := normalizeHost(rawHost)
	if host == "" {
		return Resolution{Mode: ModeUnknown}, ErrInvalidHost
	}

	switch authority {
	case r.marketingHost:
		return Resolution{Mode: ModeMarketing, Host: host}, nil
	case r.appHost:
		return Resolution{Mode: ModeApp, Host: host}, nil
	}

	switch host {
	case r.marketingName, r.baseDomain, r.wwwHost:
		return Resolution{Mode: ModeMarketing, Host: host}, nil
	case r.appHostname:
		return Resolution{Mode: ModeApp, Host: host}, nil
	}

	suffix := "." + r.baseDomain
	if !strings.HasSuffix(host, suffix) {
		return Resolution{Mode: ModeUnknown, Host: host}, ErrInvalidHost
	}

	slug := strings.TrimSuffix(host, suffix)
	if slug == "" || strings.Contains(slug, ".") {
		return Resolution{Mode: ModeUnknown, Host: host}, ErrInvalidHost
	}

	return Resolution{
		Mode: ModeTenant,
		Host: host,
		Slug: slug,
	}, nil
}

func (r *Resolver) Middleware(lookup LookupFunc, marketingURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			resolution, err := r.ResolveHost(req.Host)
			if err != nil {
				if isAPIRequest(req) {
					writeJSONError(w, http.StatusNotFound, "tenant_not_found", "Tenant not found.")
					return
				}
				http.Redirect(w, req, marketingURL, http.StatusTemporaryRedirect)
				return
			}

			if resolution.Mode == ModeTenant {
				org, lookupErr := lookup(req.Context(), resolution.Slug)
				if lookupErr != nil || org == nil {
					if isAPIRequest(req) {
						writeJSONError(w, http.StatusNotFound, "tenant_not_found", "Tenant not found.")
						return
					}
					http.Redirect(w, req, marketingURL, http.StatusTemporaryRedirect)
					return
				}
				resolution.Organization = org
			}

			ctx := context.WithValue(req.Context(), tenantContextKey, resolution)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context) Resolution {
	if value, ok := ctx.Value(tenantContextKey).(Resolution); ok {
		return value
	}
	return Resolution{Mode: ModeUnknown}
}

func normalizeHost(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(value)
	if err == nil {
		value = host
	}
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeAuthority(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.ToLower(value)
}

func isAPIRequest(req *http.Request) bool {
	return strings.HasPrefix(req.URL.Path, "/api/")
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    code,
		"message": message,
	})
}
