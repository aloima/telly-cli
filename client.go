package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func ReadStdin(stdin chan string, isQuitted chan bool) {
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

			stdin <- text[:len(text)-1]
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

const HELP = ("use \"help\" for helping\nuse \"quit\" or \"exit\" to quit\nuse \"clear\" to clear screen\nresponse format is `(type) value`\n")

func SplitCommand(input string) []string {
	var result []string
	var escaping bool
	var value string

	for _, c := range input {
		if c == ' ' && !escaping {
			if value != "" {
				result = append(result, value)
				value = ""
			}
		} else if c == '"' {
			if !escaping {
				escaping = true
			} else {
				escaping = false

				if value != "" {
					result = append(result, value)
					value = ""
				}
			}
		} else {
			value += string(c)
		}
	}

	if value != "" {
		result = append(result, value)
	}

	return result
}

func StartClient(host string, port int) {
	url := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", url)

	if err != nil {
		log.Fatal(err)
	}

	stdin := make(chan string, 1)
	isQuitted := make(chan bool, 1)

	fmt.Print(HELP)

	go ReadStdin(stdin, isQuitted)

	for {
		fmt.Print(">> ")
		input := <-stdin

		switch input {
		case "quit", "exit":
			isQuitted <- true
			close(stdin)
			fmt.Println("quitted")
			return

		case "clear":
			fmt.Print("\033[H\033[2J")

		case "help":
			fmt.Print(HELP)

		default:
			arr := SplitCommand(input)
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
