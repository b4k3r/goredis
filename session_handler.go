package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"strings"

	"github.com/google/shlex"
	"github.com/sirupsen/logrus"
)

var ErrAuth = errors.New("(error) NOAUTH Authentication required.")

type sessionHandler struct {
	conn       io.ReadWriteCloser
	logger     *logrus.Entry
	storage    map[string]string
	authorized bool
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
		case "AUTH":
			err = s.authCommand(tokens[1:])
		case "SET":
			err = s.setCommand(tokens[1:])
		case "GET":
			err = s.getCommand(tokens[1:])
		case "PING":
			err = s.pingCommand(tokens[1:])
		case "QUIT":
			return
		default:
			_, err = fmt.Fprintln(s.conn, "(error) unknown command")
		}

		if err == ErrAuth {
			_, err = fmt.Fprintln(s.conn, err)
		}

		// http://teaching.idallen.com/cst8165/08w/notes/eof_handling.txt
		if err == io.EOF {
			s.logger.Errorf("Cannot print message to client %v", err)
			break
		}
	}
}

func (s *sessionHandler) authCommand(tokens []string) error {
	if len(tokens) != 1 {
		_, err := fmt.Fprintln(s.conn, "(error) ERR wrong number of arguments for 'auth' command")
		return err
	}

	if len(password) == 0 {
		_, err := fmt.Fprintln(s.conn, "(error) ERR Client sent AUTH, but no password is set")
		return err
	}

	// Secure compare?
	if password == tokens[0] {
		s.authorized = true
		_, err := fmt.Fprintln(s.conn, "OK")
		return err
	}

	s.authorized = false
	_, err := fmt.Fprintln(s.conn, "(error) ERR invalid password")
	return err
}

func (s *sessionHandler) setCommand(tokens []string) error {
	if len(tokens) != 2 {
		_, err := fmt.Fprintln(s.conn, "(error) ERR wrong number of arguments for 'set' command")
		return err
	}

	if !s.authorized {
		return ErrAuth
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

	if !s.authorized {
		return ErrAuth
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

	if !s.authorized {
		return ErrAuth
	}

	_, err := fmt.Fprintln(s.conn, "PONG")
	return err
}
