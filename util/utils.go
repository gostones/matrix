package util

import (
	"bytes"
	"fmt"
	"github.com/jpillora/backoff"
	"net"
	"os"
	"time"
)

func BackoffDuration() func(int) {
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    60 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	return func(rc int) {
		secs := b.Duration()
		fmt.Fprintf(os.Stdout, "rc: %v sleeping %v\n", rc, secs)
		time.Sleep(secs)
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
