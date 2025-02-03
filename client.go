package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
)

func ReadStdin(stdin chan string, quit chan bool) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-quit:
			return

		default:
			exist := scanner.Scan()

			if exist {
				stdin <- scanner.Text()
			}
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

func InterpretValue(at *int64, value string, depth int) (ValueType, string) {
	*at += 1

	switch value[*at-1] {
	case '$':
		var length int64

		if value[*at] == '-' && value[*at+1] == '1' {
			*at += 4
			return Null, ""
		} else {
			for {
				if value[*at] != '\r' {
					length = (int64(value[*at]) - '0') + (length * 10)
				} else {
					*at += 2
					break
				}

				*at += 1
			}

			response := value[*at:(*at + length)]
			*at += (length + 2)

			if strings.Contains(response, "\n") {
				return BulkString, response
			} else {
				return BulkString, fmt.Sprintf("\"%s\"", response)
			}
		}

	case '+':
		start := *at

		for {
			if value[*at] != '\r' {
				*at += 1
			} else {
				*at += 2
				return BasicString, value[start:(*at - 2)]
			}
		}

	case '-':
		start := *at

		for {
			if value[*at] != '\r' {
				*at += 1
			} else {
				*at += 2
				return ErrorString, value[start:(*at - 2)]
			}
		}

	case '_':
		return Null, ""

	case ':':
		start := *at

		for {
			if value[*at] != '\r' {
				*at += 1
			} else {
				*at += 2
				return Number, value[start:(*at - 2)]
			}
		}

	case '*':
		start := *at
		var count int

		for {
			if value[*at] != '\r' {
				*at += 1
			} else {
				count, _ = strconv.Atoi(value[start:*at])
				*at += 2
				break
			}
		}

		if count == 0 {
			return Array, "(empty array)"
		}

		i := 2
		var arr []string
		offset := int(math.Log10(float64(count)))

		{
			_, subValue := InterpretValue(at, value, depth)
			arr = append(arr, fmt.Sprintf("%s1) %s", strings.Repeat(" ", offset), subValue))
		}

		for i <= count {
			_, subValue := InterpretValue(at, value, depth+offset+3)
			arr = append(arr, fmt.Sprintf("%s%d) %s", strings.Repeat(" ", depth+offset-int(math.Log10(float64(i)))), i, subValue))
			i += 1
		}

		return Array, strings.Join(arr, "\n")

	default:
		return Unknown, ""
	}
}

const HELP = ("use \"help\" for helping\nuse \"quit\" or \"exit\" to quit\nuse \"clear\" to clear screen\nresponse format is `(type)\\nvalue`\n")

func SplitCommand(input string) []string {
	var result []string
	var escaping bool
	var builder strings.Builder

	for _, c := range input {
		if c == ' ' && !escaping {
			if builder.Len() != 0 {
				result = append(result, builder.String())
				builder.Reset()
			}
		} else if c == '"' {
			if !escaping {
				escaping = true
			} else {
				escaping = false

				if builder.Len() != 0 {
					result = append(result, builder.String())
					builder.Reset()
				}
			}
		} else {
			builder.WriteRune(c)
		}
	}

	if builder.Len() != 0 {
		result = append(result, builder.String())
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
	quit := make(chan bool, 1)

	fmt.Print(HELP)

	go ReadStdin(stdin, quit)

	for {
		fmt.Print(">> ")
		input := <-stdin

		switch input {
		case "quit", "exit":
			quit <- true
			close(stdin)
			fmt.Println("quitted")
			return

		case "clear":
			fmt.Print("\033[H\033[2J")

		case "help":
			fmt.Print(HELP)

		case "":
			continue

		default:
			arr := SplitCommand(input)
			var values string

			for _, data := range arr {
				values += fmt.Sprintf("$%d\r\n%s\r\n", len(data), data)
			}

			fmt.Fprintf(conn, "*%d\r\n%s", len(arr), values)
			buf := make([]byte, 1024)
			var response string
			var builder strings.Builder

			for {
				n, err := conn.Read(buf)

				if err != nil {
					log.Fatal(err)
				}

				builder.Write(buf)

				if n != 1024 {
					response = builder.String()
					break
				}
			}

			at := new(int64)
			valueType, value := InterpretValue(at, response, 0)

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
