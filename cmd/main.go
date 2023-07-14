package main

import (
	"log"
	"net"

	"github.com/yangzuo0621/zero2prod-go/app"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := app.Run(listener); err != nil {
		log.Fatalf("failed to start app: %v", err)
	}
}
