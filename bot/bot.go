package bot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gostones/matrix/util"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"time"
)

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

	delay := 5 * time.Second

	go func() {
		me := fmt.Sprintf(`/me {"name": "%v", "type": "bot", "addr": "%v", "status": "on" , "uuid": "%v"}`, c.User, addr, c.UUID)
		fmt.Println("Sending me detail: ", me)

		for {
			_, err := send(in, me)

			if err == nil {
				break
			}

			time.Sleep(delay)
		}

		active = true
	}()

	//in.Write([]byte("/theme mono\r\n"))

	//go func() {
	//	for {
	//		in.Write([]byte("/motd\r\n"))
	//		time.Sleep(*check)
	//	}
	//}()

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
			cmd := &Command{
				cm,
			}

			if !active {
				continue
			}

			// //
			// if cm.Type == "me" && cm.Msg["name"] == c.User && cm.Msg["uuid"] == c.UUID {

			// 	go func() {
			// 		link := fmt.Sprintf(`/link %v:%v {"remote":"%v", "local":"%v"}`, cm.From, c.To, c.Remote, c.Local)
			// 		fmt.Printf("establishing link: %v\n", link)

			// 		for {
			// 			_, err := send(in, link)

			// 			if err == nil {
			// 				break
			// 			}

			// 			time.Sleep(delay)
			// 		}
			// 		fmt.Printf("link established: %v\n", link)

			// 	}()

			// 	continue
			// }

			//
			if cmd.Msg == nil || cmd.Msg["cmd"] == "" {
				continue
			}

			robot, err := getRobot(cm.Msg["cmd"])
			if err != nil {
				continue
			}

			execute := func() string {
				defer func() string {
					if r := recover(); r != nil {
						fmt.Println("Recovered in f", r)
						return fmt.Sprintf(`{"error": "%v"}`, r)
					}
					return `{"error": "unknown"}`
				}()
				return robot.Run(cmd)
			}

			if resp := execute(); resp != "" {
				//TODO check from
				send(in, fmt.Sprintf(`/msg %v %v`, cm.From, resp))
			}
		}

		//if strings.Contains(line, " "+*user+": ") {
		//	cmd, err := parseLine(line)
		//	if err == nil {
		//		robot, err := getRobot(cmd.Command)
		//		if err != nil {
		//			continue
		//		}
		//
		//		if !active {
		//			continue
		//		}
		//
		//		if response := robot.Run(cmd); response != "" {
		//			reply(in, fmt.Sprintf("%s %s", cmd.From, response))
		//		}
		//	}
		//}
	}

	return errors.New("ERROR")
}

//
//func parseLine(line string) (*robots.Command, error) {
//	fields := strings.Fields(line)
//
//	if len(fields) < 4 {
//		return nil, errors.New("not enough fields in line")
//	}
//
//	fromFields := strings.Split(fields[1], controlCodeString)
//	if len(fromFields) < 2 {
//		return nil, errors.New("not enough fields in line")
//	}
//	from := fromFields[1]
//
//	if len(fields) < 4 {
//		return nil, errors.New("not enough fields in line")
//	}
//
//	command := strings.TrimRight(fields[3], "\a")
//
//	args := []string{}
//
//	if len(fields) > 4 {
//		for _, f := range fields[4:] {
//			args = append(args, strings.TrimRight(f, "\a"))
//		}
//	}
//
//	if active {
//		fmt.Printf("%#v\n", args)
//	}
//
//	cmd := robots.Command{
//		From:    from,
//		Command: command,
//		Args:    args,
//	}
//
//	return &cmd, nil
//}

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

func getRobot(command string) (Robot, error) {
	if robotInitFunction, ok := Robots[command]; ok {
		return robotInitFunction(), nil
	}

	return nil, fmt.Errorf("unknown robot: %s", command)
}
