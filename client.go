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

func InterpretValue(value string) string {
	switch value[0] {
	case '$':
		var length int64
		var at int64 = 1

		for {
			if value[at] != '\r' {
				length = (int64(value[at]) - '0') + (length * 10)
			} else {
				at += 2 // pass '\n'
				break
			}

			at += 1
		}

		return fmt.Sprintf("(bulk string)\n\"%s\"", value[at:(at+length)])

	case '+':
		var at int64 = 1

		for {
			if value[at] != '\r' {
				at += 1
			} else {
				return fmt.Sprintf("(basic string)\n%s", value[1:at])
			}
		}

	case '-':
		var at int64 = 1

		for {
			if value[at] != '\r' {
				at += 1
			} else {
				return fmt.Sprintf("(error)\n%s", value[1:at])
			}
		}

	default:
		return ""
	}
}

const HELP = ("use \"help\" for helping\nuse \"quit\" to quit\nresponse format is `(type) value`\n")

func StartClient(host string, port int) {
	defer func() {
		exec.Command("stty", "-f", "/dev/tty", "echo").Run()
	}()

	url := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", url)

	if err != nil {
		log.Fatal(err)
	}

	stdin := make(chan string, 1)
	quit := make(chan bool, 1)

	fmt.Print(HELP)

	go ReadStdin(stdin, quit)

	for {
		fmt.Print(">> ")
		input := <-stdin

		if input == "quit" {
			quit <- true
			close(stdin)
			fmt.Println("quitted")
			break
		} else if input == "help" {
			fmt.Print(HELP)
		} else {
			// TODO: string escaping
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

			fmt.Println(InterpretValue(response))
		}
	}
}
