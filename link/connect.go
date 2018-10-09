package link

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
)

// Connect to remote service
func Connect(c *Config) error {
	var active = false
	var remotePort = -1
	var svcUser = ""

	var timeout = 120 * 1000 // 2 min

	fmt.Fprintf(os.Stdout, "Connect proxy: %v server: %v\n", c.Proxy, c.URL)

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

	//
	greet := func() error {
		meMsg := fmt.Sprintf(`/me {"name": "%v", "type": "link", "addr": "%v", "status": "on" , "uuid": "%v"}`, c.User, addr, c.UUID)
		whoisMsg := fmt.Sprintf(`/whois %v`, c.Service.Name)

		fmt.Println("Sending greetings ...")
		fmt.Println(meMsg)
		fmt.Println(whoisMsg)

		if !active {
			_, err := send(in, meMsg)
			fmt.Printf("greet: %v\n", err)

			if err != nil {
				return err
			}
			active = true
		}
		if active && remotePort == -1 {
			_, err := send(in, whoisMsg)
			fmt.Printf("greet: %v\n", err)
			if err != nil {
				return err
			}
		}

		return nil
	}

	//
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

					//notify
					c.Ready <- true
				}

				if cm.Type == "presence" && cm.Msg["who"] == svcUser {
					fmt.Printf("service user presence who: %v status: %v\r\n", svcUser, cm.Msg["status"])

					if cm.Msg["status"] == "left" {
						remotePort = -1
						svcUser = ""

						//
						c.Ready <- false
					}
				}
			}
		}

		return scanner.Err()
	}

	//
	fmt.Printf("Connect greeting\n")

	err = greet()
	if err != nil {
		return err
	}

	fmt.Printf("Connect handle msg\n")

	err = handle()

	return err
}
