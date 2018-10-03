package link

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gostones/matrix/util"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"strconv"
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
}

// ChatMessage format
type ChatMessage struct {
	Type string            `json:"type"`
	To   string            `json:"to"`
	From string            `json:"from"`
	Msg  map[string]string `json:"msg"`
}

// Serve
func Serve(c *Config) error {
	var active = false
	var remotePort = -1

	fmt.Fprintf(os.Stdout, "RP Service, proxy: %v server: %v\n", c.Proxy, c.URL)

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	conn, err := dial(addr, c.User)
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

	// var delay = 5 * time.Second

	var count = 0

	var doneChan = make(chan bool, 1)
	var tickChan = time.NewTicker(time.Second * 5).C
	var timeChan = time.NewTimer(time.Second * 120).C

	greet := func() {
		me := fmt.Sprintf(`/me {"name": "%v", "type": "link", "addr": "%v", "status": "on" , "uuid": "%v"}`, c.User, addr, c.UUID)
		svcMsg := fmt.Sprintf(`/svc {"host_port":"%v", "uuid":"%v"}`, c.Service.HostPort, c.UUID)

		if !active {
			_, err := send(in, me)
			fmt.Printf("greet: %v\n", err)

			if err == nil {
				active = true
			}
		}
		if active {
			_, err := send(in, svcMsg)
			fmt.Printf("greet: %v\n", err)
		}
	}

	svc := func() {
		for {
			select {
			case <-timeChan:
				fmt.Println("Timer expired")
				panic("Timeout")
			case <-tickChan:
				fmt.Println("Ticker ticked")
				greet()
				count++
				fmt.Printf("error: %v count: %v\n", err, count)
			case <-doneChan:
				fmt.Printf("Service started, count: %v\n", count)
				return
			}
		}
	}

	// go func() {
	// 	me := fmt.Sprintf(`/me {"name": "%v", "type": "link", "addr": "%v", "status": "on" , "uuid": "%v"}`, c.User, addr, c.UUID)
	// 	fmt.Println("Sending me detail: ", me)

	// 	for {
	// 		_, err := send(in, me)

	// 		if err == nil {
	// 			break
	// 		}

	// 		time.Sleep(delay)
	// 	}


	// 	//
	// 	svc()
	// }()

	go svc()

	//
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		line := scanner.Text()
		if err != nil {
			return err
		}
		fmt.Println("Got: ", line)

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
					doneChan <- true
					tunRPC(c, remotePort)
					panic("Failed to reverse proxy")
				}
			}
		}
	}

	return errors.New("ERROR")
}

func send(in io.WriteCloser, s string) (int, error) {
	return in.Write([]byte(s + "\r\n"))
}

func dial(addr, user string) (*ssh.Client, error) {
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
