/*
|--------------------------------------------------------------------------
| Body Parser Configuration
|--------------------------------------------------------------------------
|
| Limits and allowed content types for incoming request bodies.
|
*/

package config

var BodyParser BodyParserConfig

type BodyParserConfig struct {
	JSONLimit      string
	FormLimit      string
	MultipartLimit string
	AllowedTypes   []string
}

func loadBodyParser() {
	BodyParser = BodyParserConfig{
		JSONLimit:      env("BODY_JSON_LIMIT", "1mb"),
		FormLimit:      env("BODY_FORM_LIMIT", "1mb"),
		MultipartLimit: env("BODY_MULTIPART_LIMIT", "10mb"),
		AllowedTypes:   []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"},
	}
}
