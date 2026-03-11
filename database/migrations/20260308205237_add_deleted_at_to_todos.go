package migrations

import "gorm.io/gorm"

// AddDeletedAtToTodos adds deleted_at for GORM soft delete.
type AddDeletedAtToTodos struct{}

// TableName returns the migration name for tracking.
func (m *AddDeletedAtToTodos) TableName() string {
	return "todos"
}

// Up adds the deleted_at column.
func (m *AddDeletedAtToTodos) Up(db *gorm.DB) error {
	return db.Exec("ALTER TABLE todos ADD COLUMN deleted_at DATETIME").Error
}

// Down removes the deleted_at column (SQLite doesn't support DROP COLUMN easily; no-op).
func (m *AddDeletedAtToTodos) Down(db *gorm.DB) error {
	return nil
}
