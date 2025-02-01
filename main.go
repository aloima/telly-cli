package main

import (
	"fmt"
	"os"
)

func main() {
	switch len(os.Args) {
	case 1:
		StartClient()

	case 2:
		arg := os.Args[1]

		if arg == "help" {
			// TODO
			os.Exit(0)
		} else {
			fmt.Println("invalid argument")
			os.Exit(1)
		}

	default:
		fmt.Println("invalid count of arguments")
	}
}
