package util

import (
	"fmt"
	"sync"
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

func TestTimedBoom(t *testing.T) {

	ticker := func() {
		fmt.Print(" . ")
	}

	t.Logf("Test timeout")

	Timed(
		100, ticker,
		1000, func() {
			fmt.Println("Boom timeout ...")

		},
		100, 500, func() error {
			select {}
		})
}

func TestTimedComplete(t *testing.T) {

	ticker := func() {
		fmt.Print(" . ")
	}

	t.Logf("Test complete")

	Timed(
		100, ticker,
		30000, func() {
			fmt.Println("Boom complete ...")
		},
		100, 500, func() error {
			time.Sleep(500 * time.Millisecond)

			fmt.Println("Return complete ...")
			return nil
		})
}

func TestTimedRunning(t *testing.T) {

	ticker := func() {
		fmt.Print(" . ")
	}

	t.Logf("Test keep running")

	wg := sync.WaitGroup{}
	wg.Add(1)

	Timed(
		100, ticker,
		500, func() {
			fmt.Println("Boom keep running ...")
			//terminate
			select {
			case <-time.After(2 * time.Second):
				wg.Done()
			}
			wg.Wait()
		},
		100, 500, func() error {
		loop:
			{
				fmt.Print(" X ")
				time.Sleep(1 * time.Second)

			}
			goto loop
		})
}
