package models

import "github.com/CodeSyncr/nimbus/database"

// Todo embeds the base model (ID, timestamps).
type Todo struct {
	database.Model
	Title string
	Done  bool
}
