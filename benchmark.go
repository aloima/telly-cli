package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func PrintBenchmarkTypes() {
	out := `
List of benchmark types
  ping
	  runs "PING" command
  time
	  runs "TIME" command
`

	fmt.Println(out[1:len(out)-1])
}

func create_client(id int, host string, port int) net.Conn {
	addr := net.JoinHostPort(host, fmt.Sprint(port))
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Fatalf("client %d cannot connect", id)
	}

	return conn
}

func benchmark(client net.Conn, in string) {
	buf := make([]byte, 1024)
	fmt.Fprint(client, in)

	for {
		n, err := client.Read(buf)

		if err != nil {
			log.Fatal(err)
		}

		if n != 1024 {
			break
		}
	}
}

func StartBenchmark(host string, port int, types string, client_count int, request_count uint64) {
	var clients []net.Conn

	for i := range client_count {
		clients = append(clients, create_client(i, host, port))
	}

	if types == "" {
		log.Fatal("types are empty")
	}

	type_values := strings.Split(types, ",")

	for _, t := range type_values {
		var in string

		switch t {
		case "ping":
			in = "*1\r\n$4\r\nPING\r\n"

		default:
			log.Fatalf("invalid type: %s\n", t)
		}

		for i := range client_count {
			go benchmark(clients[i], in)
		}

		// TODO: add waiter for clients
	}
}
