package migrations

import (
	"github.com/CodeSyncr/nimbus"
	"github.com/CodeSyncr/nimbus/database/schema"
)

// CreateTodos migration — Laravel-inspired schema style.
type CreateTodos struct {
	schema.BaseSchema
}

// TableName returns the migration name for tracking.
func (m *CreateTodos) TableName() string {
	return "todos"
}

// Up creates the todos table.
func (m *CreateTodos) Up(db *nimbus.DB) error {
	return schema.New(db).CreateTable("todos", func(table *schema.Table) {
		table.Increments("id")
		table.String("title", 255)
		table.Boolean("done").Default("0")
		table.Timestamps()
	})
}

// Down drops the todos table.
func (m *CreateTodos) Down(db *nimbus.DB) error {
	return schema.New(db).DropTable("todos")
}
