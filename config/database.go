/*
|--------------------------------------------------------------------------
| Database Configuration
|--------------------------------------------------------------------------
|
| Connection settings for the primary database. Supported drivers:
| sqlite, postgres, mysql.
|
*/

package config

var Database DatabaseConfig

type DatabaseConfig struct {
	Driver string
	DSN    string
}

func loadDatabase() {
	driver := cfg("database.driver", "sqlite")
	var dsn string

	switch driver {
	case "postgres", "pg":
		dsn = cfg("database.dsn", "")
		if dsn == "" {
			dsn = "host=" + cfg("database.host", "localhost") +
				" port=" + cfg("database.port", "5432") +
				" user=" + cfg("database.user", "postgres") +
				" password=" + cfg("database.password", "") +
				" dbname=" + cfg("database.database", "nimbus") +
				" sslmode=disable"
		}
	case "mysql":
		dsn = cfg("database.dsn", "")
		if dsn == "" {
			dsn = cfg("database.user", "root") + ":" + cfg("database.password", "") +
				"@tcp(" + cfg("database.host", "localhost") + ":" + cfg("database.port", "3306") + ")/" +
				cfg("database.database", "nimbus") + "?charset=utf8mb4&parseTime=True"
		}
	default:
		dsn = cfg("database.dsn", "database.sqlite")
	}

	Database = DatabaseConfig{
		Driver: driver,
		DSN:    dsn,
	}
}
