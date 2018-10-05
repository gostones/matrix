package util

import (
	"fmt"
	"testing"
	"time"
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

func TestTimed(t *testing.T) {
	duration := 1 * 100
	timeout := 10 * 100

	min := 100
	max := 500

	ticker := func() {
		fmt.Println("Ticking ....")
	}

	boomer := func() {
		//panic("timeout")
		fmt.Println("Boom boom boom")
	}

	fn := func() error {
		for i := 0; i < 2; i++ {
			fmt.Printf("%d ", i)
			time.Sleep(100 * time.Millisecond)
		}
		return fmt.Errorf("error")
	}

	Timed(duration, ticker, timeout, boomer, min, max, fn)

	//time.Sleep(1 * time.Hour)
 }