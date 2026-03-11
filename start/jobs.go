package start

import (
	"github.com/CodeSyncr/nimbus/queue"

	"nimbus-starter/app/jobs"
)

// RegisterQueueJobs registers all queue jobs for the application.
// Called from bin/server.go via queue.Boot.
func RegisterQueueJobs() {
	queue.Register(&jobs.SendWelcomeEmail{})
}