// Command: nimbus db:migrate
package main

import (
	"fmt"
	"os"

	"github.com/CodeSyncr/nimbus/database"

	"nimbus-starter/config"
	"nimbus-starter/database/migrations"
)

func main() {
	config.Load()
	db, err := database.Connect(config.Database.Driver, config.Database.DSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
		os.Exit(1)
	}
	migrator := database.NewMigrator(db, migrations.All())
	if err := migrator.Up(); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Migrations completed.")
}
