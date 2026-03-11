/*
|--------------------------------------------------------------------------
| Rate Limiter Configuration
|--------------------------------------------------------------------------
|
| Default rate-limiting rules.
|
*/

package config

import "time"

var Limiter LimiterConfig

type LimiterConfig struct {
	Enabled  bool
	Requests int
	Window   time.Duration
	KeyFunc  string // "ip" | "user" | "custom"
}

func loadLimiter() {
	Limiter = LimiterConfig{
		Enabled:  envBool("RATE_LIMIT_ENABLED", true),
		Requests: envInt("RATE_LIMIT_REQUESTS", 100),
		Window:   time.Duration(envInt("RATE_LIMIT_WINDOW_SECONDS", 60)) * time.Second,
		KeyFunc:  env("RATE_LIMIT_KEY", "ip"),
	}
}
