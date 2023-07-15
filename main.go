package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	_ "github.com/lib/pq"
	"github.com/yangzuo0621/zero2prod-go/app"
)

func main() {
	configuration, err := app.GetConfiguration(".")
	if err != nil {
		log.Fatalf("Failed to read configuration: %v", err)
	}

	db, err := sql.Open("postgres", configuration.Database.ConnectionString())
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	address := fmt.Sprintf("127.0.0.1:%d", configuration.ApplicationPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := app.Run(listener, db); err != nil {
		log.Fatalf("failed to start app: %v", err)
	}
}
