package link

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gostones/matrix/util"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const maxInputLength int = 1024

// Service
type Service struct {
	Name     string
	HostPort string
	Port     int
}

// Config
type Config struct {
	Host  string
	Port  int
	Proxy string

	URL   string //matrix
	MPort int    //matrix

	UUID string
	User string

	Service *Service

	Ready chan bool
}

// ChatMessage format
type ChatMessage struct {
	Type string            `json:"type"`
	To   string            `json:"to"`
	From string            `json:"from"`
	Msg  map[string]string `json:"msg"`
}

// Serve starts reverse proxy service
func Serve(c *Config) error {
	//
	var active = false
	var remotePort = -1

	var timeout = 120 * 1000 //  min

	fmt.Printf("RP Service, proxy: %v server: %v\n", c.Proxy, c.URL)

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	conn, err := dial(addr, c.User, timeout)

	if err != nil {
		return err
	}

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	in, err := session.StdinPipe()
	if err != nil {
		return err
	}

	out, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	err = session.RequestPty("xterm", 1, maxInputLength, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	greet := func() error {
		meMsg := fmt.Sprintf(`/me {"name": "%v", "type": "link", "addr": "%v", "status": "on" , "uuid": "%v"}`, c.User, addr, c.UUID)
		svcMsg := fmt.Sprintf(`/svc {"host_port":"%v", "uuid":"%v"}`, c.Service.HostPort, c.UUID)

		fmt.Println("Sending greetings ...")
		fmt.Println(meMsg)
		fmt.Println(svcMsg)

		if !active {
			_, err := send(in, meMsg)
			fmt.Printf("greet: %v\n", err)

			if err != nil {
				return err
			}
			active = true
		}
		if active && remotePort == -1 {
			_, err := send(in, svcMsg)
			fmt.Printf("greet: %v\n", err)
			if err != nil {
				return err
			}
		}

		return nil
	}

	handle := func() error {
		scanner := bufio.NewScanner(out)

		for scanner.Scan() {
			line := scanner.Text()

			fmt.Println("Got: ", line)
			if strings.HasPrefix(line, "/") {
				continue
			}

			cm := ChatMessage{}
			err := json.Unmarshal([]byte(line), &cm)
			if err != nil {
				fmt.Printf("Json error (TODO): %v\n", line)
			} else {
				if !active {
					continue
				}

				//
				if cm.Msg == nil {
					continue
				}

				if cm.Type == "system" && cm.Msg["type"] == "port" && cm.Msg["uuid"] == c.UUID {
					if cm.Msg["error"] == "" {
						remotePort = parseInt(cm.Msg["remote_port"], -1)
						fmt.Printf("response remote_port: %v\n", remotePort)

						//
						go func() {
							tunRPC(c, remotePort)

							c.Ready <- false
						}()

						//notify
						c.Ready <- true
					}
				}
			}
		}

		return scanner.Err()
	}

	//
	fmt.Printf("RP Service, greeting\n")

	err = greet()
	if err != nil {
		return err
	}

	fmt.Printf("RP Service, handling msg\n")

	err = handle()
	return err
}

func send(in io.WriteCloser, s string) (int, error) {
	return in.Write([]byte(s + "\r\n"))
}

func dial(addr, user string, timeout int) (*ssh.Client, error) {
	key, err := util.MakeKey()
	if err != nil {
		return nil, err
	}

	return ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Duration(timeout) * time.Second,
	})
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
