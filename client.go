package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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

type ValueType string

const (
	Null        ValueType = "null"
	ErrorString ValueType = "error"
	BasicString ValueType = "basic string"
	BulkString  ValueType = "bulk string"
	Number      ValueType = "number"
	Array       ValueType = "array"
	Unknown     ValueType = "unknown"
)

func InterpretValue(value string) (ValueType, string) {
	switch value[0] {
	case '$':
		var length int64
		var at int64 = 1

		if value[1] == '-' && value[2] == '1' {
			return Null, ""
		} else {
			for {
				if value[at] != '\r' {
					length = (int64(value[at]) - '0') + (length * 10)
				} else {
					at += 2 // pass '\n'
					break
				}

				at += 1
			}

			response := value[at:(at + length)]

			if strings.Contains(response, "\n") {
				return BulkString, response
			} else {
				return BulkString, fmt.Sprintf("\"%s\"", response)
			}
		}

	case '+':
		var at int64 = 1

		for {
			if value[at] != '\r' {
				at += 1
			} else {
				return BasicString, value[1:at]
			}
		}

	case '-':
		var at int64 = 1

		for {
			if value[at] != '\r' {
				at += 1
			} else {
				return ErrorString, value[1:at]
			}
		}

	case '_':
		return Null, ""

	case ':':
		var at int64 = 1

		for {
			if value[at] != '\r' {
				at += 1
			} else {
				return Number, value[1:at]
			}
		}

	default:
		return Unknown, ""
	}
}

const HELP = ("use \"help\" for helping\nuse \"quit\" or \"exit\" to quit\nuse \"clear\" to clear screen\nresponse format is `(type)\\nvalue`\n")

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

			valueType, value := InterpretValue(response)

			if value == "" {
				fmt.Printf("(%s)\n", valueType)
			} else if valueType == BulkString {
				if strings.Contains(value, "\n") {
					fmt.Printf("(%s)\n%s", valueType, value)
				} else {
					fmt.Printf("(%s)\n%s\n", valueType, value)
				}
			} else {
				fmt.Printf("(%s)\n%s\n", valueType, value)
			}
		}
	}
}
