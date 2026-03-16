/*
|--------------------------------------------------------------------------
| Shield Configuration
|--------------------------------------------------------------------------
|
| Shield protects your application by setting a variety of HTTP
| headers. It guards against clickjacking, XSS, content-type
| sniffing, and other common attacks.
|
| The CSRF sub-config defines token-based CSRF protection.
| Set ExceptPaths to bypass CSRF on specific routes (e.g. webhooks,
| API endpoints that use token auth instead).
|
*/

package config

import (
	"net/http"
	"time"
)

var Shield ShieldConfig

// ShieldConfig mirrors shield.Config in the Nimbus framework.
type ShieldConfig struct {
	// ── Content Security ────────────────────────────────

	// ContentTypeNosniff adds X-Content-Type-Options: nosniff
	ContentTypeNosniff bool

	// XSSProtection sets the X-XSS-Protection header.
	// Modern browsers ignore this; set to "0" to disable.
	XSSProtection string

	// FrameGuard sets X-Frame-Options. Values: "DENY", "SAMEORIGIN".
	FrameGuard string

	// ── HSTS ────────────────────────────────────────────

	HSTS HSTSConfig

	// ── Referrer & DNS ──────────────────────────────────

	// ReferrerPolicy controls the Referrer-Policy header.
	// Common values: "no-referrer", "strict-origin-when-cross-origin"
	ReferrerPolicy string

	// DNSPrefetchControl sets X-DNS-Prefetch-Control.
	DNSPrefetchControl bool

	// DownloadOptions adds X-Download-Options: noopen (IE).
	DownloadOptions bool

	// ── Cross-Origin Policies ───────────────────────────

	PermittedCrossDomainPolicies string
	CrossOriginOpenerPolicy      string
	CrossOriginResourcePolicy    string
	CrossOriginEmbedderPolicy    string

	// ── CSP ─────────────────────────────────────────────

	CSP CSPShieldConfig

	// ── CSRF ────────────────────────────────────────────

	CSRF CSRFConfig
}

type HSTSConfig struct {
	// Enabled turns on Strict-Transport-Security header.
	// Only enable this in production with TLS.
	Enabled bool

	// MaxAge is how long browsers should remember the HSTS policy.
	MaxAge time.Duration

	// IncludeSubdomains extends HSTS to all subdomains.
	IncludeSubdomains bool

	// Preload adds the preload directive (submit to hstspreload.org).
	Preload bool
}

type CSPShieldConfig struct {
	// Enabled turns on Content-Security-Policy headers.
	Enabled bool

	// ReportOnly uses Content-Security-Policy-Report-Only instead.
	ReportOnly bool
}

type CSRFConfig struct {
	// Enabled turns on CSRF protection. Set to false for
	// pure-API apps that use token authentication.
	Enabled bool

	// CookieName is the name of the CSRF cookie.
	CookieName string

	// HeaderName is the header that carries the CSRF token.
	HeaderName string

	// FieldName is the form field name for the CSRF token.
	FieldName string

	// MaxAge is the CSRF cookie lifetime in seconds.
	MaxAge int

	// Secure marks the cookie as HTTPS-only.
	Secure bool

	// SameSite controls the SameSite attribute.
	SameSite http.SameSite

	// Path restricts the cookie to a specific path.
	Path string

	// Domain restricts the cookie to a specific domain.
	Domain string

	// HttpOnly prevents client-side JS from reading the cookie.
	HttpOnly bool

	// ExceptPaths skips CSRF validation for these paths.
	// Useful for webhooks, external API endpoints, etc.
	ExceptPaths []string

	// RotateToken generates a new token on every request.
	RotateToken bool
}

func loadShield() {
	isProd := env("APP_ENV", "development") == "production"

	Shield = ShieldConfig{
		ContentTypeNosniff: true,
		XSSProtection:      "0",
		FrameGuard:         "SAMEORIGIN",

		HSTS: HSTSConfig{
			Enabled:           isProd,
			MaxAge:            365 * 24 * time.Hour,
			IncludeSubdomains: true,
			Preload:           false,
		},

		ReferrerPolicy:               "strict-origin-when-cross-origin",
		DNSPrefetchControl:           false,
		DownloadOptions:              true,
		PermittedCrossDomainPolicies: "none",
		CrossOriginOpenerPolicy:      "same-origin",
		CrossOriginResourcePolicy:    "same-origin",
		CrossOriginEmbedderPolicy:    "",

		CSP: CSPShieldConfig{
			Enabled:    false,
			ReportOnly: false,
		},

		CSRF: CSRFConfig{
			Enabled:     envBool("CSRF_ENABLED", true),
			CookieName:  env("CSRF_COOKIE", "__nimbus_csrf"),
			HeaderName:  "X-CSRF-Token",
			FieldName:   "_csrf",
			MaxAge:      envInt("CSRF_MAX_AGE", 86400),
			Secure:      isProd,
			SameSite:    http.SameSiteLaxMode,
			Path:        "/",
			Domain:      "",
			HttpOnly:    true,
			ExceptPaths: []string{},
			RotateToken: false,
		},
	}
}
