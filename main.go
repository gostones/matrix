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
		server    - server mode
		bot       - service worker
		cli       - control agent
		service   - link service
		connect   - connect to service
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
	case "cli":
		client(args)
	case "service":
		linkService(args)
	case "connect":
		linkConnect(args)
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

const rps = `
[common]
bind_port = %v
`
const matrixPort = 2022
const rpsPort = 8000

func server(args []string) {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)

	//tunnel port
	bind := flags.Int("bind", parseInt(os.Getenv("PORT"), 8080), "")

	//chat port
	port := flags.Int("port", parseInt(os.Getenv("MATRIX_PORT"), matrixPort), "")
	ident := flags.String("identity", os.Getenv("MATRIX_IDENTITY"), "")

	v := flags.Bool("verbose", false, "")

	rport := flags.Int("rps", rpsPort, "")
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
		sleep(fmt.Errorf("error: %d", rc))
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

	go tunnel.TunClient(*proxy, *url, fmt.Sprintf("localhost:%v:localhost:%v", lport, *port))

	sleep := util.BackoffDuration()

	for {
		rc := bot.Server(&cfg)
		sleep(fmt.Errorf("error: %d", rc))
	}
}

func linkConnect(args []string) {
	flags := flag.NewFlagSet("connect", flag.ContinueOnError)

	port := flags.Int("port", parseInt(os.Getenv("MATRIX_PORT"), 2022), "")
	url := flags.String("url", os.Getenv("MATRIX_URL"), "")
	proxy := flags.String("proxy", "", "")

	name := flags.String("name", "", "")

	toName := flags.String("service", "", "remote service name")
	listenPort := flags.Int("listen", util.FreePort(), "local port for exposing remote service")

	flags.Parse(args)

	if *url == "" {
		usage()
	}

	if *toName == "" {
		usage()
	}

	//
	lport := util.FreePort()
	user := *name
	if user == "" {
		user = fmt.Sprintf("connect%v", lport)
	}

	cfg := link.Config{
		Host:  "localhost",
		Port:  lport,
		Proxy: *proxy,
		URL:   *url,
		UUID:  uuid.New().String(),
		User:  user,

		Service: &link.Service{
			Name: *toName,
			Port: *listenPort,
		},
	}
	//
	fmt.Fprintf(os.Stdout, "link service: %v user: %v\n", cfg, user)

	go tunnel.TunClient(*proxy, *url, fmt.Sprintf("localhost:%v:localhost:%v", lport, *port))

	link.Connect(&cfg)
	fmt.Println("Failed to connect")
}

func linkService(args []string) {
	flags := flag.NewFlagSet("service", flag.ContinueOnError)

	port := flags.Int("port", parseInt(os.Getenv("MATRIX_PORT"), 2022), "")
	url := flags.String("url", os.Getenv("MATRIX_URL"), "")

	proxy := flags.String("proxy", "", "")
	name := flags.String("name", "", "")

	toHostPort := flags.String("hostport", "", "reverse proxy service host:port")

	flags.Parse(args)

	if *url == "" {
		usage()
	}

	if *toHostPort == "" {
		usage()
	}

	lport := util.FreePort()
	user := *name
	if user == "" {
		user = fmt.Sprintf("svc%v", lport)
	}

	cfg := &link.Config{
		Host:  "localhost",
		Port:  lport,
		Proxy: *proxy,
		URL:   *url,
		MPort: *port,
		UUID:  uuid.New().String(),
		User:  user,

		Service: &link.Service{
			HostPort: *toHostPort,
			Port:     rpsPort,
		},
	}

	//
	fmt.Fprintf(os.Stdout, "Staring link service: %v user: %v\n", cfg, cfg.User)
	go tunnel.TunClient(*proxy, *url, fmt.Sprintf("localhost:%v:localhost:%v", lport, *port))

	link.Serve(cfg)
	fmt.Println("Failed to start service")
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
