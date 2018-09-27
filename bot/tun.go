package bot

import (
	"fmt"
	"github.com/jpillora/chisel/client"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// TunBot creates a tunnel
type TunBot struct {
	url   string
	proxy string
}

func init() {
	RegisterRobot("tun", func() (robot Robot) {
		return &TunBot{
			url:   MatrixUrl,
			proxy: ProxyUrl,
		}
	})
}

// Run executes a command
func (b TunBot) Run(c *Command) string {
	// if len(c.Args) != 2 {
	// 	return "missing ports: local remote"
	// }

	hostPort := strings.Split(c.Msg["host_port"], ":")
	//lhost := hostPort[0]

	lport, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	rport, err := strconv.Atoi(c.Msg["remote_port"])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	go b.tun(lport, rport)

	return fmt.Sprintf("Started local: %v remote: %v", lport, rport)
}

// Description describes what the robot does
func (b TunBot) Description() string {
	return "local_port remote_port"
}

func (b TunBot) tun(lport, rport int) {

	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, rport)

	fmt.Fprintf(os.Stdout, "remote: %v proxy: %v url: %v\n", remote, b.proxy, b.url)

	b.tunClient(b.proxy, b.url, remote)
}

func (b TunBot) tunClient(proxy string, url string, remote string) {

	keepalive := time.Duration(12 * time.Second)

	c, err := chclient.NewClient(&chclient.Config{
		Fingerprint: "",
		Auth:        "",
		KeepAlive:   keepalive,
		HTTPProxy:   proxy,
		Server:      url,
		Remotes:     []string{remote},
	})

	c.Debug = true

	defer c.Close()
	if err = c.Run(); err != nil {
		log.Println(err)
	}
}
