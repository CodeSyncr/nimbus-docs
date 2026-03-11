package seeders

import "github.com/CodeSyncr/nimbus/database"

// All returns seeders in run order. Add new seeders when you run nimbus make:seeder.
func All() []database.Seeder {
	return []database.Seeder{
		// &UserSeeder{},
	}
}
