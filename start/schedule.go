package start

import (
	"github.com/CodeSyncr/nimbus/schedule"
)

// RegisterSchedule defines scheduled tasks for the application.
// This is invoked by bin.RunSchedule via "nimbus schedule:run".
//
// Example:
//
//	s.EveryMinute("health-check", func(ctx context.Context) error {
//	    return healthService.Ping(ctx)
//	})
//	s.Daily("03:00", "nightly-cleanup", func(ctx context.Context) error {
//	    return cleanupService.Run(ctx)
//	})
func RegisterSchedule(s *schedule.Scheduler) {
	// No tasks by default.
}
