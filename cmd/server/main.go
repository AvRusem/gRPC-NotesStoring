package main

import (
	"fmt"
	"os"

	"cu.ru/internal/servers"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: server (required <host:port>) (optional <sql_connection_string>)")
		os.Exit(1)
	}

	addr := os.Args[1]
	dsn := ""
	if len(os.Args) >= 3 {
		dsn = os.Args[2]
	}

	servers.StartServer(addr, dsn)
}
