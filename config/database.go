/*
|--------------------------------------------------------------------------
| Database Configuration
|--------------------------------------------------------------------------
|
| Connection settings for SQL and NoSQL databases. Nimbus supports
| multiple named connections — register them in start/kernel.go.
|
| Supported SQL drivers:  sqlite, postgres, mysql
| Supported NoSQL drivers: mongodb
|
| Multi-DB Example (in start/kernel.go):
|
|   database.ConnectAll([]database.ConnectionConfig{
|       {Name: "default", Driver: config.Database.Driver, DSN: config.Database.DSN},
|       {Name: "analytics", Driver: config.Database.Connections["analytics"].Driver, DSN: config.Database.Connections["analytics"].DSN},
|   })
|
| NoSQL Example (in start/kernel.go):
|
|   mongo, _ := nosql.ConnectMongo(ctx, nosql.MongoConfig{
|       URI:      config.Database.MongoURI,
|       Database: config.Database.MongoDatabase,
|   })
|   nosql.Register("mongo", mongo)
|
*/

package config

var Database DatabaseConfig

type DatabaseConnectionConfig struct {
	Driver string
	DSN    string
}

type DatabaseConfig struct {
	// Primary SQL connection
	Driver string
	DSN    string

	// Additional named SQL connections
	Connections map[string]DatabaseConnectionConfig

	// MongoDB / NoSQL
	MongoURI      string
	MongoDatabase string
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

	// Build additional named connections from env
	connections := make(map[string]DatabaseConnectionConfig)

	// Example: ANALYTICS_DB_DRIVER=postgres, ANALYTICS_DB_DSN=...
	analyticsDriver := cfg("analytics_db.driver", "")
	if analyticsDriver != "" {
		connections["analytics"] = DatabaseConnectionConfig{
			Driver: analyticsDriver,
			DSN:    cfg("analytics_db.dsn", ""),
		}
	}

	Database = DatabaseConfig{
		Driver:        driver,
		DSN:           dsn,
		Connections:   connections,
		MongoURI:      cfg("mongo.uri", ""),
		MongoDatabase: cfg("mongo.database", ""),
	}
}
