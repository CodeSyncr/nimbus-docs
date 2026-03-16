/*
|--------------------------------------------------------------------------
| CORS Configuration
|--------------------------------------------------------------------------
|
| Cross-Origin Resource Sharing (CORS) controls which external
| domains can make requests to your API. This configuration sets
| the Access-Control-* response headers.
|
| ── Configuring Allowed Origins ─────────────────────────
|
| AllowOrigins accepts a slice of domain strings:
|   - []string{"*"}                                → allow all origins
|   - []string{"https://app.example.com"}          → specific domain
|   - []string{"https://app.example.com", "..."}   → multiple domains
|
| WARNING: When AllowCredentials is true, browsers reject the
| literal "*" origin. Nimbus automatically reflects the
| requesting origin instead of sending "*".
|
| ── Configuring Allowed Methods ─────────────────────────
|
| AllowMethods defines which HTTP methods are permitted in
| cross-origin requests. The browser's preflight request checks
| the Access-Control-Request-Method header against this list.
|
| ── Configuring Allowed Headers ─────────────────────────
|
| AllowHeaders controls which request headers are permitted.
| Use []string{"*"} to allow all headers, or list specific ones:
|   []string{"Content-Type", "Authorization", "X-Requested-With"}
|
| ── Exposing Response Headers ───────────────────────────
|
| By default browsers only expose basic headers to JavaScript.
| ExposeHeaders lets you whitelist additional response headers:
|   []string{"X-Request-Id", "X-RateLimit-Remaining"}
|
| ── Credentials ─────────────────────────────────────────
|
| Enable AllowCredentials when your frontend sends cookies or
| the Authorization header. Without this, browsers strip
| credentials from cross-origin requests.
|
| ── Caching Preflight ───────────────────────────────────
|
| MaxAge (seconds) tells browsers how long to cache preflight
| responses, reducing repeated OPTIONS requests. Set to -1 to
| send the header but disable caching.
|
| See: /docs/cors
|
*/

package config

var CORS CORSConfig

type CORSConfig struct {
	// Enabled controls whether CORS headers are sent.
	Enabled bool

	// AllowOrigins is the list of origins that may access
	// your API. Use ["*"] to allow all origins.
	AllowOrigins []string

	// AllowMethods lists HTTP methods allowed for cross-origin
	// requests.
	AllowMethods []string

	// AllowHeaders lists the request headers allowed in
	// cross-origin requests. Use ["*"] to allow all.
	AllowHeaders []string

	// ExposeHeaders lists response headers that browsers may
	// access from JavaScript.
	ExposeHeaders []string

	// AllowCredentials enables cookies, Authorization headers,
	// and TLS client certificates in cross-origin requests.
	AllowCredentials bool

	// MaxAge is how long (seconds) browsers cache preflight
	// responses. Set -1 to disable caching.
	MaxAge int
}

func loadCORS() {
	CORS = CORSConfig{
		Enabled:          envBool("CORS_ENABLED", true),
		AllowOrigins:     []string{env("CORS_ORIGIN", "*")},
		AllowMethods:     []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{},
		AllowCredentials: envBool("CORS_CREDENTIALS", false),
		MaxAge:           envInt("CORS_MAX_AGE", 86400),
	}
}
