/*
|--------------------------------------------------------------------------
| Rate Limiter Configuration
|--------------------------------------------------------------------------
|
| The rate limiter protects your application from abuse by
| throttling the number of requests a client can make within a
| given time window.
|
| ── Key Function ────────────────────────────────────────
|
| KeyFunc determines how clients are identified:
|   - "ip"     → throttle by client IP address (default)
|   - "user"   → throttle by authenticated user ID
|   - "custom" → provide your own key extractor
|
| ── Stores ──────────────────────────────────────────────
|
| Store controls where rate limit counters are kept:
|   - "memory" → in-process (single instance only)
|   - "redis"  → distributed across multiple instances
|
| ── Global vs Route-Level ──────────────────────────────
|
| This config applies globally. For route-level limits, use
| named middleware:
|
|   app.Router.Post("/api/login", handler).
|       Use(middleware.RateLimit(5, time.Minute, nil))
|
| See: /docs/rate-limiting
|
*/

package config

import "time"

var Limiter LimiterConfig

type LimiterConfig struct {
	// Enabled toggles the global rate limiter.
	Enabled bool

	// Requests is the max number of requests allowed per window.
	Requests int

	// Window is the time period that Requests applies to.
	Window time.Duration

	// KeyFunc determines how clients are identified.
	// Values: "ip" | "user" | "custom"
	KeyFunc string

	// Store controls where counters are stored.
	// Values: "memory" | "redis"
	Store string

	// RedisURL is used when Store is "redis".
	RedisURL string

	// Headers controls whether X-RateLimit-* headers are sent.
	Headers bool

	// BlockDuration is how long a client is blocked after
	// exceeding the limit. Defaults to the Window duration.
	BlockDuration time.Duration
}

func loadLimiter() {
	window := time.Duration(envInt("RATE_LIMIT_WINDOW_SECONDS", 60)) * time.Second

	Limiter = LimiterConfig{
		Enabled:       envBool("RATE_LIMIT_ENABLED", true),
		Requests:      envInt("RATE_LIMIT_REQUESTS", 100),
		Window:        window,
		KeyFunc:       env("RATE_LIMIT_KEY", "ip"),
		Store:         env("RATE_LIMIT_STORE", "memory"),
		RedisURL:      env("REDIS_URL", ""),
		Headers:       envBool("RATE_LIMIT_HEADERS", true),
		BlockDuration: time.Duration(envInt("RATE_LIMIT_BLOCK_SECONDS", 0)) * time.Second,
	}
	if Limiter.BlockDuration == 0 {
		Limiter.BlockDuration = Limiter.Window
	}
}
