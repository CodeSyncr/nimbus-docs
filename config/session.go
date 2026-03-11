/*
|--------------------------------------------------------------------------
| Session Configuration
|--------------------------------------------------------------------------
|
| Cookie-based session settings.
|
*/

package config

var Session SessionConfig

type SessionConfig struct {
	Driver     string // cookie | memory
	CookieName string
	MaxAge     int
	HttpOnly   bool
	Secure     bool
	SameSite   string // strict | lax | none
}

func loadSession() {
	Session = SessionConfig{
		Driver:     env("SESSION_DRIVER", "cookie"),
		CookieName: env("SESSION_COOKIE", "nimbus_session"),
		MaxAge:     envInt("SESSION_MAX_AGE", 86400*7),
		HttpOnly:   true,
		Secure:     env("APP_ENV", "development") == "production",
		SameSite:   "lax",
	}
}
