package validators

import "github.com/CodeSyncr/nimbus/validation"

// Todo validates todo input. Implements SchemaProvider for fluent validation.
type Todo struct {
	Title   string
	Content string
}

// Rules returns VineJS-style validation schema.
func (v *Todo) Rules() validation.Schema {
	return validation.Schema{
		"title": validation.String().Required().Min(1).Max(255).Trim(),
	}
}

// Validate runs the schema rules and returns an error if validation fails.
func (v *Todo) Validate() error {
	return validation.ValidateStruct(v)
}
