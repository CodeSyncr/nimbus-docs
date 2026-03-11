package jobs

import (
	"context"
	"fmt"

	"github.com/CodeSyncr/nimbus/queue"
)

// SendWelcomeEmail is a demo job used in the nimbus-starter app
// to showcase queueing and Horizon.
type SendWelcomeEmail struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
}

// Handle is called by the queue worker when the job is processed.
func (j *SendWelcomeEmail) Handle(ctx context.Context) error {
	fmt.Printf("[queue] sending welcome email to %s (user_id=%d)\n", j.Email, j.UserID)
	return nil
}

// Failed is called when the job exhausts all retry attempts.
func (j *SendWelcomeEmail) Failed(ctx context.Context, err error) {
	fmt.Printf("[queue] failed welcome email to %s: %v\n", j.Email, err)
}

var (
	_ queue.Job       = (*SendWelcomeEmail)(nil)
	_ queue.FailedJob = (*SendWelcomeEmail)(nil)
)

