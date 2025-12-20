package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
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

func benchmark(completed_req_count *uint64, req_count uint64, client net.Conn, in string) {
	buf := make([]byte, 1024)

	for {
		if *completed_req_count >= req_count {
			break
		}

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

		*completed_req_count += 1
	}
}

func StartBenchmark(host string, port int, types string, client_count int, req_count uint64) {
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

		case "time":
			in = "*1\r\n$4\r\nTIME\r\n"

		default:
			log.Fatalf("invalid type: %s\n", t)
		}

		completed_req_count := uint64(0)

		for i := range client_count {
			go benchmark(&completed_req_count, req_count, clients[i], in)
		}

		start := time.Now()

		for {
			if completed_req_count >= req_count {
				elapsed := time.Since(start)
				fmt.Printf("====== %s ======\n", strings.ToUpper(t))
				fmt.Printf("%d parallel clients\n", client_count)
				fmt.Printf("%d requests completed in %.3f seconds\n", completed_req_count, elapsed.Seconds())
				fmt.Printf("%.3f requests per second\n\n", float64(completed_req_count) / elapsed.Seconds())
				break
			}
		}
	}
}
