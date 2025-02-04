package main

import (
	"flag"
)

func main() {
	host := flag.String("host", "127.0.0.1", "hostname or ip address")
	port := flag.Int("port", 6379, "port number")

	flag.Parse()

	StartClient(*host, *port)
}
