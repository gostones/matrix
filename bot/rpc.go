package bot

import (
	"fmt"
	"github.com/gostones/matrix/rp"
	"github.com/gostones/matrix/util"
	"github.com/jpillora/chisel/client"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

//server_port, instance, service_host, service_port, remote_port
var rpc = `
[common]
server_addr = localhost
server_port = %v
http_proxy =

[rpc%v]
type = tcp
local_ip = %v
local_port = %v
remote_port = %v
`

// RpcBot starts reverse proxy client
type RpcBot struct {
	url   string
	proxy string
}

func init() {
	RegisterRobot("rpc", func() (robot Robot) {
		return &RpcBot{
			url:   MatrixUrl,
			proxy: ProxyUrl,
		}
	})
}

// Run executes a command
func (b RpcBot) Run(c *Command) string {
	// if len(c.Args) != 3 {
	// 	return "missing args: host port remote_port"
	// }

	hostPort := strings.Split(c.Msg["host_port"], ":")
	shost := hostPort[0]

	sport, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	rport, err := strconv.Atoi(c.Msg["remote_port"])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	go b.tun(shost, sport, rport)

	return fmt.Sprintf("Started service: %v remote: %v", sport, rport)
}

// Description describes what the robot does
func (b RpcBot) Description() string {
	return "service_port remote_port"
}

func (b RpcBot) tun(shost string, sport int, rport int) {
	lport := util.FreePort()

	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, 8000)

	go b.tunClient(b.proxy, b.url, remote)

	fmt.Fprintf(os.Stdout, "service: %v remote: %v proxy: %v url: %v\n", sport, remote, b.proxy, b.url)

	sleep := util.BackoffDuration()

	for {
		rc := rp.Client(fmt.Sprintf(rpc, lport, rport, shost, sport, rport))
		if rc == 0 {
			return
		}
		sleep(rc)
	}
}

func (b RpcBot) tunClient(proxy string, url string, remote string) {

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
