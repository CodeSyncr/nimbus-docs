/*
|--------------------------------------------------------------------------
| Application Configuration
|--------------------------------------------------------------------------
|
| The app key is used for encrypting cookies and generating signed
| URLs. Keep it secret and never commit it to version control.
|
*/

package config

var App AppConfig

type AppConfig struct {
	Name string
	Env  string // development | production | test
	Port int
	Host string
	Key  string

	HTTP HTTPConfig
}

type HTTPConfig struct {
	AllowMethodSpoofing bool
	Cookie               CookieConfig
}

type CookieConfig struct {
	Domain   string
	Path     string
	MaxAge   int // seconds
	HttpOnly bool
	Secure   bool
	SameSite string // strict | lax | none
}

func loadApp() {
	App = AppConfig{
		Name: cfg("app.name", "nimbus-starter"),
		Env:  cfg("app.env", "development"),
		Port: cfgInt("app.port", 3333),
		Host: cfg("app.host", "0.0.0.0"),
		Key:  cfg("app.key", ""),
		HTTP: HTTPConfig{
			AllowMethodSpoofing: false,
			Cookie: CookieConfig{
				Domain:   "",
				Path:     "/",
				MaxAge:   7200,
				HttpOnly: true,
				Secure:   cfg("app.env", "development") == "production",
				SameSite: "lax",
			},
		},
	}
}
