package util

import (
	"bytes"
	"fmt"
	"github.com/jpillora/backoff"
	"net"
	"os"
	"time"
)

func BackoffDuration() func(error) {
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    15 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	return func(rc error) {
		secs := b.Duration()

		fmt.Fprintf(os.Stdout, "rc: %v sleeping %v\n", rc, secs)
		time.Sleep(secs)
		if secs.Nanoseconds() >= b.Max.Nanoseconds() {
			b.Reset()
		}
	}
}

func FreePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func MacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}

func HttpProxyEnv() string {
	p := os.Getenv("http_proxy")
	if p == "" {
		p = os.Getenv("HTTP_PROXY")
	}
	if p == "" {
		return ""
	}
	return p
}

// Timed runs fn and boom if timeout
func Timed(timeout int, boom func(), min, max int, fn func() error) {
	if timeout < 0 {
		timeout = 60 * 1000
	}
	if min < 0 {
		min = 1 * 1000
	}
	if max < 0 {
		max = 30 * 1000
	}

	done := make(chan bool)

	go func() {
		b := &backoff.Backoff{
			Min:    time.Duration(min) * time.Millisecond,
			Max:    time.Duration(max) * time.Millisecond,
			Factor: 2,
			Jitter: false,
		}

		count := 0
		for {
			fmt.Println("Calling fn ...")
			if err := fn(); err != nil {
				count++
				fmt.Printf("fn %d %v\n", count, err)

				d := b.Duration()

				time.Sleep(d)
				if d.Nanoseconds() >= b.Max.Nanoseconds() {
					b.Reset()
				}
				continue
			}
			done <- true
			fmt.Println("Returning from fn ...")
			return
		}
	}()

	for {
		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			fmt.Println("boom boom boom!")
			if boom != nil {
				boom()
			}
			return
		case <-done:
			fmt.Println("Done!")
			return
		}
	}
}

// //Timed runs fn per backoff in [min, max] and calls ticker at duration and boomer if timeout in milli second
// func Timed(duration int, ticker func(), timeout int, boomer func(), min, max int, fn func() error) {

// 	tick := time.Tick(time.Duration(duration) * time.Millisecond)
// 	boom := time.After(time.Duration(timeout) * time.Millisecond)

// 	done := make(chan bool, 1)

// 	go func() {
// 		b := &backoff.Backoff{
// 			Min:    time.Duration(min) * time.Millisecond,
// 			Max:    time.Duration(max) * time.Millisecond,
// 			Factor: 2,
// 			Jitter: false,
// 		}

// 		count := 0
// 		for {
// 			fmt.Println("Calling fn ...")
// 			if err := fn(); err != nil {
// 				count++
// 				fmt.Printf("fn %d %v\n", count, err)

// 				d := b.Duration()

// 				time.Sleep(d)
// 				if d.Nanoseconds() >= b.Max.Nanoseconds() {
// 					b.Reset()
// 				}
// 				continue
// 			}
// 			done <- true
// 			fmt.Println("Returning from fn ...")
// 			return
// 		}
// 	}()

// 	for {
// 		select {
// 		case <-done:
// 			fmt.Println("Done!")
// 			return
// 		case <-tick:
// 			fmt.Print(" - ")
// 			if ticker != nil {
// 				ticker()
// 			}
// 		case <-boom:
// 			fmt.Println("Boom boom boom!")
// 			if boomer != nil {
// 				boomer()
// 			}
// 			return
// 		}
// 	}
// }
