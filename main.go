package main

import (
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/gostones/matrix/bot"
	"github.com/gostones/matrix/chat"
	"github.com/gostones/matrix/link"
	"github.com/gostones/matrix/rp"
	"github.com/gostones/matrix/ssh"
	"github.com/gostones/matrix/tunnel"
	"github.com/gostones/matrix/util"
	"os"
	"strconv"
)

//
var help = `
	Usage: matrix [command] [--help]

	Commands:
		server - server mode
		bot    - service worker
		link   - link service
		cli    - control agent
`

//
func main() {

	flag.Bool("help", false, "")
	flag.Bool("h", false, "")
	flag.Usage = func() {}
	flag.Parse()

	args := flag.Args()

	subcmd := ""
	if len(args) > 0 {
		subcmd = args[0]
		args = args[1:]
	}

	//
	switch subcmd {
	case "server":
		server(args)
	case "bot":
		botService(args)
	case "link":
		linkService(args)
	case "cli":
		client(args)
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, help)
	os.Exit(1)
}

//func genUser(rand int) string {
//	return fmt.Sprintf("u_%v_%v", strings.Replace(util.MacAddr(), ":", "", -1), rand)
//}

var rps = `
[common]
bind_port = %v
`
var matrixPort = 2022

func server(args []string) {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)

	//tunnel port
	bind := flags.Int("bind", parseInt(os.Getenv("PORT"), 8080), "")

	//chat port
	port := flags.Int("port", parseInt(os.Getenv("MATRIX_PORT"), matrixPort), "")
	ident := flags.String("identity", os.Getenv("MATRIX_IDENTITY"), "")

	v := flags.Bool("verbose", false, "")

	rport := flags.Int("rps", 8000, "")
	sport := flags.Int("ssh", 8022, "")

	flags.Parse(args)

	//
	args = []string{}
	args = append(args, "--bind", fmt.Sprintf(":%v", *port))

	if *ident == "" {
		*ident = "host_key"
		util.RsaKeyPair(*ident)
	}
	args = append(args, "--identity", *ident)

	if *v {
		args = append(args, "-v")
	}

	//
	go chat.Server(args)

	go rp.Server(fmt.Sprintf(rps, *rport))
	go ssh.Server(*sport, "bash")

	tunnel.TunServer(fmt.Sprintf("%v", *bind))
}

func client(args []string) {
	flags := flag.NewFlagSet("cli", flag.ContinueOnError)

	lport := util.FreePort()

	port := flags.Int("port", parseInt(os.Getenv("MATRIX_PORT"), matrixPort), "")
	ident := flags.String("identity", os.Getenv("MATRIX_IDENTITY"), "")
	url := flags.String("url", os.Getenv("MATRIX_URL"), "")
	proxy := flags.String("proxy", "", "")
	user := flags.String("name", fmt.Sprintf("cli%v", lport), "")

	// to := flags.String("to", "", "target servcie name")
	// remote := flags.String("remote", "", "host:port")
	// local := flags.String("local", "", ":port")

	flags.Parse(args)

	if *url == "" {
		usage()
	}

	if *proxy == "" {
		*proxy = os.Getenv("http_proxy")
	}

	if *ident == "" {
		*ident = "host_key"
		util.RsaKeyPair(*ident)
	}

	//
	fmt.Fprintf(os.Stdout, "local: %v remote: %v\n", lport, *port)

	//
	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, *port)
	go tunnel.TunClient(*proxy, *url, remote)

	args = []string{"--p", fmt.Sprintf("%v", lport), "--i", *ident, *user + "@localhost"}
	sleep := util.BackoffDuration()
	for {
		rc := ssh.Client(args)
		if rc == 0 {
			os.Exit(0)
		}
		sleep(rc)
	}
}

func botService(args []string) {
	flags := flag.NewFlagSet("bot", flag.ContinueOnError)

	lport := util.FreePort()

	port := flags.Int("port", parseInt(os.Getenv("MATRIX_PORT"), 2022), "")
	url := flags.String("url", os.Getenv("MATRIX_URL"), "")
	proxy := flags.String("proxy", "", "")
	user := flags.String("name", fmt.Sprintf("bot%v", lport), "")

	flags.Parse(args)

	if *url == "" {
		usage()
	}

	cfg := bot.Config{
		Host:  "localhost",
		Port:  lport,
		Proxy: *proxy,
		URL:   *url,
		UUID:  uuid.New().String(),
		User:  *user,
	}
	//
	fmt.Fprintf(os.Stdout, "local: %v user: %v\n", lport, user)

	// remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, *port)
	go tunnel.TunClient(*proxy, *url, fmt.Sprintf("localhost:%v:localhost:%v", lport, *port))

	sleep := util.BackoffDuration()

	for {
		rc := bot.Server(&cfg)
		sleep(rc)
	}
}

func linkService(args []string) {
	flags := flag.NewFlagSet("link", flag.ContinueOnError)

	lport := util.FreePort()

	port := flags.Int("port", parseInt(os.Getenv("MATRIX_PORT"), 2022), "")
	url := flags.String("url", os.Getenv("MATRIX_URL"), "")
	proxy := flags.String("proxy", "", "")
	user := flags.String("name", fmt.Sprintf("link%v", lport), "")

	toName := flags.String("link-name", "", "remote service name")
	toHostPort := flags.String("link-hostport", "", "remote service host:port")
	fromPort := flags.Int("link-port", util.FreePort(), "")

	flags.Parse(args)

	if *url == "" {
		usage()
	}

	if *toName == "" || *toHostPort == "" {
		usage()
	}

	rpPort := 11022

	cfg := link.Config{
		Host:  "localhost",
		Port:  lport,
		Proxy: *proxy,
		URL:   *url,
		UUID:  uuid.New().String(),
		User:  *user,

		Service: &link.Service{
			Name:     *toName,
			HostPort: *toHostPort,
			Port:     rpPort,
		},
	}
	//
	fmt.Fprintf(os.Stdout, "link local: %v user: %v\n", lport, user)

	//service link
	go tunnel.TunClient(*proxy, *url, fmt.Sprintf("localhost:%v:localhost:%v", *fromPort, rpPort))

	//chat
	go tunnel.TunClient(*proxy, *url, fmt.Sprintf("localhost:%v:localhost:%v", lport, *port))

	// sleep := util.BackoffDuration()

	// for {
	// 	rc := link.Serve(&cfg)
		
	// 	sleep(rc)
	// }

	link.Serve(&cfg)
}

func parseInt(s string, v int) int {
	if s == "" {
		return v
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		i = v
	}
	return i
}
