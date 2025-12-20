package main

import (
	"flag"
)

func main() {
	host := flag.String("host", "127.0.0.1", "hostname or ip address")
	port := flag.Int("port", 6379, "port number")
	benchmark := flag.Bool("benchmark", false, "enables benchmark mode")
	clients := flag.Int("clients", 10, "client count for benchmark")
	requests := flag.Uint64("requests", 1000, "total request count for benchmark")
	types := flag.String("types", "", "benchmark types")
	type_list := flag.Bool("typelist", false, "list benchmark types")

	flag.Parse()

	if *type_list {
		PrintBenchmarkTypes()
	} else if *benchmark {
		StartBenchmark(*host, *port, *types, *clients, *requests)
	} else {
		StartClient(*host, *port)
	}
}
