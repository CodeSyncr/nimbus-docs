/*
|--------------------------------------------------------------------------
| Static Files Configuration
|--------------------------------------------------------------------------
|
| Serve static assets from the public/ directory.
|
*/

package config

var Static StaticConfig

type StaticConfig struct {
	Enabled bool
	Root    string
	Prefix  string
	MaxAge  int
}

func loadStatic() {
	Static = StaticConfig{
		Enabled: envBool("STATIC_ENABLED", true),
		Root:    env("STATIC_ROOT", "public"),
		Prefix:  env("STATIC_PREFIX", "/public"),
		MaxAge:  envInt("STATIC_MAX_AGE", 86400),
	}
}
