/*

ssh-chat-bot

A small chatbot for ssh-chat
https://github.com/peterhellberg/ssh-chat-bot
*/
package bot

import (
	"fmt"
	"os"
)

//var (
//	user    = flag.String("n", "ssh-chat-bot", "Username")
//	owner   = flag.String("o", "peterhellberg", "Bot owner username")
//	host    = flag.String("h", "localhost", "Hostname")
//	port    = flag.Int("p", 2022, "Port")
//	verbose = flag.Bool("v", false, "Verbose output")
//	delay   = flag.Duration("d", 5*time.Second, "Delay")
//	check   = flag.Duration("c", 30*time.Second, "Duration between alive checks")
//)

// Config for bot
type Config struct {
	Host  string
	Port  int
	Proxy string

	URL  string
	UUID string
	User string
}

// Server starts bot
func Server(c *Config) int {
	ProxyUrl = c.Proxy
	MatrixUrl = c.URL

	fmt.Fprintf(os.Stdout, "Bot proxy: %v server: %v\n", c.Proxy, c.URL)

	if err := Bot(c); err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	return 0
}

//
//func usage() {
//	fmt.Fprintf(os.Stderr, "usage: ./ssh-chat-bot [-h hostname] [-v]\n\n")
//
//	if buildCommit != "" {
//		fmt.Fprintf(os.Stderr, "build: "+repoURL+"/commit/"+buildCommit+"\n\n")
//	}
//
//	fmt.Fprintf(os.Stderr, "flags:\n")
//	flag.PrintDefaults()
//	fmt.Fprintf(os.Stderr, "\n")
//	os.Exit(2)
//}
//
//func l(format string, args ...interface{}) {
//	if *verbose {
//		fmt.Printf(format+"\n", args...)
//	}
//}
