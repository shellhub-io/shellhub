package asynq

import (
	"github.com/shellhub-io/shellhub/pkg/worker"
)

// BatchTask configures a task to process a list of tasks in batches.
// Each task payload will be aggregated, separated by '\n'. Example:
//
//	func(ctx context.Context, payload []byte) error {
//	    scanner := bufio.NewScanner(bytes.NewReader(payload))
//	    scanner.Split(bufio.ScanLines)
//
//	    for scanner.Scan() {
//	        // Process each task payload
//	    }
//	}
func BatchTask() worker.TaskOption {
	return func(t *worker.Task) {
		t.Pattern += ":batch"
	}
}
