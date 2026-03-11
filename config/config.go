/*
|--------------------------------------------------------------------------
| Configuration Loader
|--------------------------------------------------------------------------
|
| config.Load() uses the type-safe Nimbus config: it loads .env and
| populates every config struct. Call this once at the top of Boot() before using any
| config value.
|
*/

package config

import nimbusconfig "github.com/CodeSyncr/nimbus/config"

// Load loads config from config.yaml, config.toml, .env and initialises all configuration structs.
func Load() {
	_ = nimbusconfig.LoadAuto()

	loadApp()
	loadDatabase()
	loadQueue()
	loadAuth()
	loadBodyParser()
	loadCache()
	loadCORS()
	loadHash()
	loadLimiter()
	loadLogger()
	loadMail()
	loadSession()
	loadStatic()
	loadStorage()
}
