// Client application that reads from keyboard add and checks operations with data to store to a bloomfilter by means of an rpc
// until pressing ctrl-c.
//
// It needs the server to be started.
//
// Example:
//
// add data1
//
// check data1
package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/devopsfaith/bloomfilter/rpc/client"
)

func main() {
	server := flag.String("server", "127.0.0.1:1234", "ip:port of the remote bloomfilter to connect to")
	flag.Parse()

	c, err := client.New(*server)
	if err != nil {
		log.Println("unable to create the rpc client:", err.Error())
		return
	}
	defer c.Close()

	in := bufio.NewReader(os.Stdin)
	for {
		line, _, err := in.ReadLine()
		if err != nil {
			log.Fatal(err)
		}

		if len(line) == 0 {
			continue
		}

		parts := strings.Split(string(line), " ")
		switch parts[0] {
		case "add":
			c.Add([]byte(strings.Join(parts[1:], " ")))
		case "check":
			ok := c.Check([]byte(strings.Join(parts[1:], " ")))
			log.Printf("%v", ok)
		default:
			log.Println("unknown command")
		}
	}
}
