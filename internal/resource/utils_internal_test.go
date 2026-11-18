package resource

import (
	"errors"
	"testing"
	"time"
)

func TestWaitStatus_BacksOffOnErrors(t *testing.T) {
	var sleeps int
	sleep = func(time.Duration) { sleeps++ }
	t.Cleanup(func() { sleep = time.Sleep })

	calls := 0
	ready := "ready"
	status, diags := waitStatus(func() (*string, error) {
		calls++
		if calls <= 2 {
			return nil, errors.New("503 service unavailable")
		}
		return &ready, nil
	}, []string{"ready"}, 5)

	if diags.HasError() || status == nil || *status != "ready" {
		t.Fatalf("status = %v, diags = %v", status, diags)
	}
	if sleeps != 2 {
		t.Errorf("sleeps = %d, want 2 (one per failed poll)", sleeps)
	}
}
