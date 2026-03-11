/*
|--------------------------------------------------------------------------
| Config Helpers
|--------------------------------------------------------------------------
|
| cfg and cfgInt read from the type-safe Nimbus config store (populated
| by config.yaml, config.toml, .env). env and envBool remain for legacy
| use where direct env access is needed.
|
*/

package config

import (
	"os"
	"strconv"

	nimbusconfig "github.com/CodeSyncr/nimbus/config"
)

func cfg(key, fallback string) string {
	return nimbusconfig.GetOrDefault(key, fallback)
}

func cfgInt(key string, fallback int) int {
	return nimbusconfig.GetOrDefault(key, fallback)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
