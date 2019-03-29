package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	t "net/textproto"
	"strings"
)

var port int

func init() {
	flag.IntVar(&port, "port", 6379, "HTTP Port")
}

func main() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	log.Printf("Listening on :%d\n", port)

	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	storage := make(map[string]string)

	for {
		// Wait for a connection.
		conn, err := l.Accept()

		if err != nil {
			log.Fatal(err)
		}

		go func(c net.Conn) {
			log.Printf("Serving %s\n", c.RemoteAddr().String())

			conn := t.NewConn(c)

			for {
				c.Write([]byte(string("redis> ")))

				r := t.NewReader(bufio.NewReader(c))
				line, err := r.ReadLine()

				if err != nil {
					log.Println(err)
				}

				tokens := strings.SplitN(line, " ", 3)

				if len(tokens) == 0 {
					continue
				}

				// Should be lexer, syntax analizator?

				switch strings.ToUpper(tokens[0]) {
				case "SET":
					// It doesn't support "" and ''

					storage[tokens[1]] = strings.TrimSpace(tokens[2])
					conn.Writer.PrintfLine("OK")
				case "GET":
					if len(tokens) != 2 {
						conn.Writer.PrintfLine("(error) syntax error")
						continue
					}

					value, ok := storage[tokens[1]]

					if ok {
						conn.Writer.PrintfLine(value)
					} else {
						conn.Writer.PrintfLine("(nil)")
					}
				case "PING":
					if len(tokens) != 1 {
						conn.Writer.PrintfLine("(error) syntax error")
					}

					conn.Writer.PrintfLine("PONG")
				case "QUIT":
					if len(tokens) != 1 {
						conn.Writer.PrintfLine("(error) syntax error")
					}

					conn.Close()
					log.Printf("%s disconnected\n", c.RemoteAddr().String())
					return
				default:
					conn.Writer.PrintfLine("(error) unknown command")
				}
			}
		}(conn)
	}
}
