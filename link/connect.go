package link

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gostones/matrix/util"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
	"sync"
)

// Connect to remote service
func Connect(c *Config) error {
	var active = false
	var remotePort = -1
	var svcUser = ""

	var timeout = 60 * 1000 // 1 min

	fmt.Fprintf(os.Stdout, "Connect proxy: %v server: %v\n", c.Proxy, c.URL)

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	var conn *ssh.Client
	var err error

	fn := func() error {
		conn, err = dial(addr, c.User, timeout)
		return err
	}

	boomer := func() {
		if conn == nil {
			panic("timeout")
		}
	}
	min := 100
	max := 3 * 1000 //3 sec

	util.Timed(0, nil, timeout, boomer, min, max, fn)

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

	//
	wg := sync.WaitGroup{}
	wg.Add(1)

	greet := func() {
		meMsg := fmt.Sprintf(`/me {"name": "%v", "type": "link", "addr": "%v", "status": "on" , "uuid": "%v"}`, c.User, addr, c.UUID)
		whoisMsg := fmt.Sprintf(`/whois %v`, c.Service.Name)

		fmt.Println("Sending greetings ...")
		fmt.Println(meMsg)
		fmt.Println(whoisMsg)

		if !active {
			_, err := send(in, meMsg)
			fmt.Printf("greet: %v\n", err)

			if err == nil {
				active = true
			}
		}
		if active && remotePort == -1 {
			_, err := send(in, whoisMsg)
			fmt.Printf("greet: %v\n", err)
		}
	}

	//
	handle := func() error {
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

				if cm.Type == "system" && cm.Msg["error"] == "" && cm.Msg["user_type"] == "service" && strings.HasPrefix(cm.Msg["name"], strings.Split(c.Service.Name, "/")[0]+"/") {

					remotePort = parseInt(cm.Msg["remote_port"], -1)
					svcUser = cm.Msg["name"]

					fmt.Printf("whois response: %v user: %v\n", cm.Msg, svcUser)
					go tun(c, remotePort)
				}

				if cm.Type == "presence" && cm.Msg["who"] == svcUser {
					fmt.Printf("service user presence who: %v status: %v\r\n", svcUser, cm.Msg["status"])

					if cm.Msg["status"] == "left" {
						remotePort = -1
						svcUser = ""

						wg.Done()
						return nil
					}
				}
			}
		}

		return errors.New("Unknown error")
	}

	//
	greet()

	util.Timed(
		0, greet,
		timeout, func() {
			if !active || remotePort == -1 || svcUser == "" {
				//panic("RP not established within set timeout")
				wg.Done()
			}
		},
		min, max, handle)

	wg.Wait()
	return errors.New("ERROR")
}
