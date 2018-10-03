package util

import (
	"fmt"
	"testing"
)

func TestBackoffDuration(t *testing.T) {
	sleep := BackoffDuration()
	count := 0
	for {
		rc := func() error {
			count++
			return fmt.Errorf("count: %v", count)
		}()
		sleep(fmt.Errorf("error: %d", rc))
	}
}
