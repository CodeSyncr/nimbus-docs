/*
|--------------------------------------------------------------------------
| CORS Configuration
|--------------------------------------------------------------------------
|
| Cross-Origin Resource Sharing settings. Tighten AllowOrigins
| for production.
|
*/

package config

var CORS CORSConfig

type CORSConfig struct {
	Enabled          bool
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

func loadCORS() {
	CORS = CORSConfig{
		Enabled:          envBool("CORS_ENABLED", true),
		AllowOrigins:     []string{env("CORS_ORIGIN", "*")},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{},
		AllowCredentials: envBool("CORS_CREDENTIALS", false),
		MaxAge:           86400,
	}
}
