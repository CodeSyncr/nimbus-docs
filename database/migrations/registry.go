package migrations

import "github.com/CodeSyncr/nimbus/database"

// All returns migrations in run order. Add new migrations here when you run make:migration.
func All() []database.Migration {
	create := &CreateTodos{}
	addDeletedAt := &AddDeletedAtToTodos{}
	return []database.Migration{
		{Name: "20260308205236_create_todos", Up: create.Up, Down: create.Down},
		{Name: "20260308205237_add_deleted_at_to_todos", Up: addDeletedAt.Up, Down: addDeletedAt.Down},
	}
}
