package link

import (
	"errors"
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

//
func tunRPC(c *Config, rport int) error {
	hostPort := strings.Split(c.Service.HostPort, ":")
	shost := hostPort[0]

	sport, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return err
	}

	//go b.tun(shost, sport, rport)
	//
	lport := util.FreePort()

	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, c.Service.Port) //c.Port == 8000

	go tunClient(c.Proxy, c.URL, remote)

	fmt.Fprintf(os.Stdout, "service: %v remote: %v proxy: %v url: %v\n", sport, remote, c.Proxy, c.URL)

	rp.Client(fmt.Sprintf(rpc, lport, rport, shost, sport, rport))

	//should not return or error

	return errors.New("failed to connect to RP server")

	// sleep := util.BackoffDuration()

	// for {
	// 	rc := rp.Client(fmt.Sprintf(rpc, lport, rport, shost, sport, rport))
	// 	if rc == 0 {
	// 		return 0
	// 	}
	// 	sleep(rc)
	// }
	// rc := rp.Client(fmt.Sprintf(rpc, lport, rport, shost, sport, rport))
	// fmt.Fprintf(os.Stdout, "service: %v remote: %v proxy: %v url: %v rc: %v \n", sport, remote, b.proxy, b.url, rc)
}

func tunClient(proxy string, url string, remote string) {

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

func tun(c *Config, rport int) error {

	remote := fmt.Sprintf("localhost:%v:localhost:%v", c.Service.Port, rport)

	fmt.Fprintf(os.Stdout, "tun remote: %v proxy: %v url: %v\n", remote, c.Proxy, c.URL)

	tunClient(c.Proxy, c.URL, remote)

	return errors.New("tunnel error")
}
