package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"strings"

	"github.com/google/shlex"
	"github.com/sirupsen/logrus"
)

type sessionHandler struct {
	conn    io.ReadWriteCloser
	logger  *logrus.Entry
	storage map[string]string
}

func (s *sessionHandler) handle() {
	defer s.logger.Info("Closed connection")
	defer s.conn.Close()

	for {
		r := textproto.NewReader(bufio.NewReader(s.conn))
		line, err := r.ReadLine()

		if err != nil {
			log.Println("Could not read line", err)
			break
		}

		// http://teaching.idallen.com/cst8165/08w/notes/eof_handling.txt
		if err == io.EOF {
			break
		}

		tokens, err := shlex.Split(line)
		if err != nil {
			fmt.Fprintln(s.conn, "Invalid argument(s)")
		}

		if len(tokens) == 0 {
			continue
		}

		switch strings.ToUpper(tokens[0]) {
		case "SET":
			err = s.setCommand(tokens[1:])
		case "GET":
			err = s.getCommand(tokens[1:])
		case "PING":
			err = s.pingCommand(tokens[1:])
		case "QUIT":
			err = s.quitCommand(tokens[1:])
		default:
			_, err = fmt.Fprintln(s.conn, "(error) unknown command")
		}

		// http://teaching.idallen.com/cst8165/08w/notes/eof_handling.txt
		if err == io.EOF {
			s.logger.Errorf("Cannot print message to client %v", err)
			break
		}
	}
}

func (s *sessionHandler) setCommand(tokens []string) error {
	if len(tokens) != 2 {
		_, err := fmt.Fprintln(s.conn, "(error) ERR wrong number of arguments for 'set' command")
		return err
	}

	s.storage[tokens[0]] = strings.TrimSpace(tokens[1])

	_, err := fmt.Fprintln(s.conn, "OK")

	return err
}

func (s *sessionHandler) getCommand(tokens []string) error {
	if len(tokens) != 1 {
		_, err := fmt.Fprintln(s.conn, "(error) ERR wrong number of arguments for 'get' command")
		return err
	}

	value, ok := s.storage[tokens[0]]

	if ok {
		_, err := fmt.Fprintf(s.conn, "\"%s\" \n", value)
		return err
	}

	_, err := fmt.Fprintln(s.conn, "(nil)")
	return err

}

func (s *sessionHandler) pingCommand(tokens []string) error {
	if len(tokens) > 0 {
		_, err := fmt.Fprintln(s.conn, "(error) ERR wrong number of arguments for 'ping' command")
		return err
	}

	_, err := fmt.Fprintln(s.conn, "PONG")
	return err
}

func (s *sessionHandler) quitCommand(tokens []string) error {
	if len(tokens) > 0 {
		_, err := fmt.Fprintln(s.conn, "(error) ERR wrong number of arguments for 'quit' command")
		return err
	}

	return io.EOF
}
