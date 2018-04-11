package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/letgoapp/go-bloomfilter/rpc/client"
)

func main() {
	c, err := client.New(":1234")
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
