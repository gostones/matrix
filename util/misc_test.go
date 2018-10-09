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

func TestTimedBoom(t *testing.T) {
	t.Log("Test timeout")

	dur := 500

	timedout := false

	Timed(
		dur, func() {
			timedout = true
			fmt.Println("should timeout ...")
		},
		0, 0, func() error {
			select {}
		})
	if !timedout {
		t.Fail()
	}
}

func TestTimedComplete(t *testing.T) {
	t.Log("Test complete")

	dur := 1000

	timedout := false

	Timed(
		dur*10, func() {
			timedout = true
			t.Log("should not be called ...")
		},
		dur/2, dur*2, func() error {
			time.Sleep(time.Duration(dur) * time.Millisecond)

			t.Log("Return complete ...")
			return nil
		})
	if timedout {
		t.Fail()
	}
}

func TestTimedRunning(t *testing.T) {

	t.Log("Test keep running")

	dur := 100

	connected := false

	Timed(
		dur, func() {
			t.Log("Boom keep running ...")
			//check if we should exit or keep running
			connected = true
		},
		dur/5, dur*5, func() error {
		loop:
			{
				fmt.Print(" X ")
				time.Sleep(time.Duration(dur) * time.Millisecond)

			}
			goto loop
		})
	if !connected {
		t.Fail()
	}
}
