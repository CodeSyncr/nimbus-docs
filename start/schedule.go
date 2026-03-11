package start

import (
	"context"

	"github.com/CodeSyncr/nimbus/scheduler"
)

// RegisterSchedule defines scheduled tasks for the application.
// This is invoked by bin.RunSchedule via "nimbus schedule:run".
//
// Example:
//
//   s.EveryHour(func(ctx context.Context) error {
//     // do work
//     return nil
//   })
//
func RegisterSchedule(s *scheduler.Scheduler) {
	// No tasks by default.
	_ = context.Background()
}

