package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
)

func ReadStdin(out chan string, isQuitted chan bool) {
	exec.Command("stty", "-f", "/dev/tty", "cbreak", "min", "1", "-echo").Run()

	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-isQuitted:
			return

		default:
			text, err := reader.ReadString('\n')

			if err != nil {
				log.Fatal(err)
			}

			out <- text[:len(text)-1]
		}
	}
}

func StartClient() {
	defer func() {
		exec.Command("stty", "-f", "/dev/tty", "echo").Run()
	}()

	conn, err := net.Dial("tcp", "localhost:6379")

	if err != nil {
		log.Fatal(err)
	}

	stdin := make(chan string, 1)
	quit := make(chan bool, 1)

	fmt.Print("press \"h\" for helping\nuse \"quit\" to quit\n")

	go ReadStdin(stdin, quit)

	for {
		fmt.Print(">> ")
		input := <-stdin

		if input == "quit" {
			quit <- true
			close(stdin)
			fmt.Println("quitted")
			break
		} else {
			arr := strings.Split(input, " ")
			var values string

			for _, data := range arr {
				values += fmt.Sprintf("$%d\r\n%s\r\n", len(data), data)
			}

			fmt.Fprintf(conn, "*%d\r\n%s", len(arr), values)
			buf := make([]byte, 1024)
			var response string

			for {
				n, err := conn.Read(buf)

				if err != nil {
					log.Fatal(err)
				}

				response += string(buf)

				if n != 1024 {
					break
				}
			}

			// TODO: parsing response
			fmt.Print(response)
		}
	}
}
