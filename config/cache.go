/*
|--------------------------------------------------------------------------
| Cache Configuration
|--------------------------------------------------------------------------
|
| Default cache driver and TTL.
|
*/

package config

import "time"

var Cache CacheConfig

type CacheConfig struct {
	Driver     string
	DefaultTTL time.Duration
}

func loadCache() {
	Cache = CacheConfig{
		Driver:     env("CACHE_DRIVER", "memory"),
		DefaultTTL: time.Duration(envInt("CACHE_TTL_MINUTES", 60)) * time.Minute,
	}
}
