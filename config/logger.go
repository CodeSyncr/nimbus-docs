/*
|--------------------------------------------------------------------------
| Logger Configuration
|--------------------------------------------------------------------------
|
| Structured logging settings (backed by uber-go/zap).
|
*/

package config

var Logger LoggerConfig

type LoggerConfig struct {
	Level  string // debug | info | warn | error
	Format string // json | console
}

func loadLogger() {
	Logger = LoggerConfig{
		Level:  env("LOG_LEVEL", "info"),
		Format: env("LOG_FORMAT", "json"),
	}
}
