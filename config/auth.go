/*
|--------------------------------------------------------------------------
| Authentication Configuration
|--------------------------------------------------------------------------
|
| Default guard and session/token settings.
|
*/

package config

var Auth AuthConfig

type AuthConfig struct {
	DefaultGuard string
	Session      SessionGuardConfig
	Token        TokenGuardConfig
}

type SessionGuardConfig struct {
	CookieName string
	MaxAge     int // seconds
}

type TokenGuardConfig struct {
	HeaderName string
	Scheme     string
	ExpiresIn  int // seconds
}

func loadAuth() {
	Auth = AuthConfig{
		DefaultGuard: env("AUTH_GUARD", "session"),
		Session: SessionGuardConfig{
			CookieName: env("SESSION_COOKIE", "nimbus_session"),
			MaxAge:     envInt("SESSION_MAX_AGE", 86400*7),
		},
		Token: TokenGuardConfig{
			HeaderName: "Authorization",
			Scheme:     "Bearer",
			ExpiresIn:  envInt("TOKEN_EXPIRES_IN", 86400),
		},
	}
}
