package link

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
	"time"
)

// Connect to remote service
func Connect(c *Config) error {
	var active = false

	var timeout = 60

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

	var delay = 5 * time.Second

	var count = 0
	var remotePort = -1
	var svcUser = ""

	var doneChan = make(chan bool, 1)
	var tickChan = time.NewTicker(time.Second * 10).C
	var timeChan = time.NewTimer(time.Second * 300).C

	whois := func() {

		whoisMsg := fmt.Sprintf(`/whois %v`, c.Service.Name)

		fmt.Printf("whois: %v count: %v\n", whoisMsg, count)

		for {
			select {
			case <-timeChan:
				fmt.Println("Timer expired")
				panic("Timedout")
			case <-tickChan:
				fmt.Println("Ticker ticked")
				_, err := send(in, whoisMsg)
				count++
				fmt.Printf("error: %v count: %v\n", err, count)
			case <-doneChan:
				fmt.Println("Done!")
				fmt.Printf("whois: %v count: %v\n", whoisMsg, count)
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
		whois()
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

			if cm.Type == "system" && cm.Msg["error"] == "" && cm.Msg["user_type"] == "service" && strings.HasPrefix(cm.Msg["name"], strings.Split(c.Service.Name, "/")[0]+"/") {

				remotePort = parseInt(cm.Msg["remote_port"], -1)
				svcUser = cm.Msg["name"]

				doneChan <- true
				fmt.Printf("whois response: %v\n", cm.Msg)
				go tun(c, remotePort)
			}

			if cm.Type == "presence" && cm.Msg["who"] == svcUser {
				fmt.Printf("service user presence who: %v status: %v\r\n", svcUser, cm.Msg["status"])

				if cm.Msg["status"] == "left" {
					whois()
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
