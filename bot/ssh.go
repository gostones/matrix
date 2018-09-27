package bot

import (
	"fmt"
	"github.com/kr/pty"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/gliderlabs/ssh"
)

// SshBot starts a ssh daemon bot
type SshBot struct {
}

func init() {
	RegisterRobot("ssh", func() (robot Robot) {
		return &SshBot{}
	})
}

// Run executes a command
func (b SshBot) Run(c *Command) string {
	// if len(c.Args) == 0 {
	// 	return "missing port"
	// }

	port, err := strconv.Atoi(c.Msg["port"])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	go b.sshd(port)

	return fmt.Sprintf("Service started at %v", port)
}

// Description describes what the robot does
func (b SshBot) Description() string {
	return "bind_port"
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func (b SshBot) sshd(port int) error {
	ssh.Handle(func(s ssh.Session) {
		cmd := exec.Command("/usr/bin/login") //exec.Command("bash")
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			cmd.Env = append(os.Environ(), fmt.Sprintf("TERM=%s", ptyReq.Term))
			f, err := pty.Start(cmd)
			if err != nil {
				panic(err)
			}
			go func() {
				for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}
			}()
			go func() {
				io.Copy(f, s) // stdin
			}()
			io.Copy(s, f) // stdout
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	})

	log.Printf("starting ssh server on port %v ...\n", port)
	err := ssh.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	return err
}
