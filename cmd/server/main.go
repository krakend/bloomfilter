// Server application that registers a bloomfilter by means of an rpc.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krakendio/bloomfilter/v2"
	"github.com/krakendio/bloomfilter/v2/rotate"
	"github.com/krakendio/bloomfilter/v2/rpc"
	"github.com/krakendio/bloomfilter/v2/rpc/server"
)

func main() {
	port := flag.Int("p", 1234, "the port to listen on")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	cfg := rpc.Config{
		rotate.Config{
			bloomfilter.Config{
				N:        10000000,
				P:        0.0000001,
				HashName: "optimal",
			},
			1000,
		},
		*port,
	}
	bf := server.New(ctx, cfg)
	for {
		select {
		case sig := <-sigs:
			log.Println("Signal intercepted:", sig)
			cancel()
			return
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
			d, _ := bf.Bloomfilter().MarshalBinary()
			log.Println("Current size of the marshalled BF:", len(d))
		}
	}
}
