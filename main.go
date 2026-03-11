/*
|--------------------------------------------------------------------------
| Nimbus Application Entry Point
|--------------------------------------------------------------------------
|
| DO NOT MODIFY THIS FILE — it is the bootstrap entrypoint for the
| Nimbus application.
|
| Configuration  → config/
| Middleware      → start/kernel.go
| Routes          → start/routes.go
| Server boot     → bin/server.go
|
| See: https://github.com/CodeSyncr/nimbus
|
*/

package main

import (
	"os"

	"nimbus-starter/bin"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		bin.RunMigrations()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		bin.RunSeeders()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "schedule:run" {
		bin.RunSchedule()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "schedule:list" {
		bin.RunScheduleList()
		return
	}
	app := bin.Boot()
	_ = app.Run()
}
