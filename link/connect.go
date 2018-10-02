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
	"time"
)

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

	URL  string
	UUID string
	User string

	Service *Service
}

// Serve
func Serve(c *Config) int {
	// ProxyUrl = c.Proxy
	// MatrixUrl = c.URL

	fmt.Fprintf(os.Stdout, "Bot proxy: %v server: %v\n", c.Proxy, c.URL)

	if err := Bot(c); err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	return 0
}

// ChatMessage format
type ChatMessage struct {
	Type string            `json:"type"`
	To   string            `json:"to"`
	From string            `json:"from"`
	Msg  map[string]string `json:"msg"`
}

var active = false

const maxInputLength int = 1024

// Bot runs the bot
func Bot(c *Config) error {
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

	var delay = 5 * time.Second

	var rpcCount = 0
	var rpcUser = ""

	var doneChan = make(chan bool, 1)
	var tickChan = time.NewTicker(time.Second * 10).C
	var timeChan = time.NewTimer(time.Second * 300).C

	rpc := func() {

		rpcMsg := fmt.Sprintf(`/msg %v {"cmd":"rpc", "host_port":"%v", "remote_port":"%v", "uuid":"%v"}`, c.Service.Name, c.Service.HostPort, c.Service.Port, c.UUID)

		fmt.Printf("rpc: %v count: %v\n", rpcMsg, rpcCount)

		// for {
		// 	_, err := send(in, rpcMsg)

		// 	if err == nil {
		// 		break
		// 	}

		// 	time.Sleep(delay)
		// }

		for {
			select {
			case <-timeChan:
				fmt.Println("Timer expired")
				panic("Timedout")
			case <-tickChan:
				fmt.Println("Ticker ticked")
				_, err := send(in, rpcMsg)
				rpcCount++
				fmt.Printf("error: %v count: %v\n", err, rpcCount)
			case <-doneChan:
				fmt.Println("Done!")
				fmt.Printf("rpc:  %v count: %v\n", rpcMsg, rpcCount)
				return
			}
		}
	}

	go func() {
		me := fmt.Sprintf(`/me {"name": "%v", "type": "link", "addr": "%v", "status": "on" , "uuid": "%v"}`, c.User, addr, c.UUID)
		fmt.Println("Sending me detail: ", me)

		for {
			_, err := send(in, me)

			if err == nil {
				break
			}

			time.Sleep(delay)
		}

		active = true

		//
		rpc()
	}()

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
			fmt.Printf("TODO: %v\r\n", line)
		} else {
			if !active {
				continue
			}

			//
			if cm.Msg == nil {
				continue
			}

			if cm.Type == "pm" && cm.Msg["error"] == "" && cm.Msg["uuid"] == c.UUID {
				rpcUser = cm.From
				doneChan <- true
				fmt.Printf("rpc reply from: %v\r\n", rpcUser)
			}

			if cm.Type == "presence" && cm.Msg["who"] == rpcUser {
				fmt.Printf("rpc presence who: %v status: %v\r\n", rpcUser, cm.Msg["status"])

				if cm.Msg["status"] == "left" {
					rpcUser = ""
					rpc()
				}
			}

			//
			// if cm.Msg == nil || cm.Msg["cmd"] == "" {
			// 	continue
			// }

			// execute := func() string {
			// 	defer func() string {
			// 		if r := recover(); r != nil {
			// 			fmt.Println("Recovered in f", r)
			// 			return fmt.Sprintf("%v", r)
			// 		}
			// 		return "ok"
			// 	}()

			// 	return ""
			// }

			// if response := execute(); response != "" {
			// 	//TODO check from
			// 	send(in, fmt.Sprintf(`/msg %v %s`, cm.From, response))
			// }
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
