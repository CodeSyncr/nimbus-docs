/*
|--------------------------------------------------------------------------
| Socialite Configuration
|--------------------------------------------------------------------------
|
| Socialite provides a simple, fluent interface for OAuth
| authentication with third-party providers like GitHub, Google,
| Discord, and Apple.
|
| Add credentials for each provider your app supports. Unused
| providers can be omitted entirely.
|
| See: /docs/socialite
|
*/

package config

var Socialite SocialiteConfig

type SocialiteConfig struct {
	// RedirectURL is the default callback URL pattern.
	// Each provider can override this in its own config.
	RedirectURL string

	// Providers maps provider names to their OAuth credentials.
	Providers map[string]SocialiteProviderConfig
}

type SocialiteProviderConfig struct {
	// ClientID is the OAuth application ID.
	ClientID string

	// ClientSecret is the OAuth application secret.
	ClientSecret string

	// RedirectURL overrides the default callback URL for this provider.
	// Leave empty to use: /auth/{provider}/callback
	RedirectURL string

	// Scopes are the OAuth permission scopes requested.
	Scopes []string
}

func loadSocialite() {
	appURL := env("APP_URL", "http://localhost:3333")

	Socialite = SocialiteConfig{
		RedirectURL: appURL + "/auth/{provider}/callback",
		Providers: map[string]SocialiteProviderConfig{
			"github": {
				ClientID:     env("GITHUB_CLIENT_ID", ""),
				ClientSecret: env("GITHUB_CLIENT_SECRET", ""),
				RedirectURL:  env("GITHUB_REDIRECT_URL", appURL+"/auth/github/callback"),
				Scopes:       []string{"user:email"},
			},
			"google": {
				ClientID:     env("GOOGLE_CLIENT_ID", ""),
				ClientSecret: env("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  env("GOOGLE_REDIRECT_URL", appURL+"/auth/google/callback"),
				Scopes:       []string{"openid", "profile", "email"},
			},
			"discord": {
				ClientID:     env("DISCORD_CLIENT_ID", ""),
				ClientSecret: env("DISCORD_CLIENT_SECRET", ""),
				RedirectURL:  env("DISCORD_REDIRECT_URL", appURL+"/auth/discord/callback"),
				Scopes:       []string{"identify", "email"},
			},
			"apple": {
				ClientID:     env("APPLE_CLIENT_ID", ""),
				ClientSecret: env("APPLE_CLIENT_SECRET", ""),
				RedirectURL:  env("APPLE_REDIRECT_URL", appURL+"/auth/apple/callback"),
				Scopes:       []string{"name", "email"},
			},
		},
	}
}
