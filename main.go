package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

var port int
var password string

func init() {
	flag.IntVar(&port, "port", 6379, "HTTP Port")
	flag.StringVar(&password, "requirepass", "", "Password")
}

func main() {
	flag.Parse()
	log := logrus.New()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
	defer l.Close()

	log.Infof("Listening on %d", port)

	if len(password) > 0 {
		log.Info("Server protected by password")
	}

	// TODO: Add mutex
	storage := make(map[string]string)

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Fatalf("Could not accept connection %v", err)
		}

		logger := log.WithField("remote", conn.RemoteAddr())
		logger.Infoln("Serving")

		go (&sessionHandler{conn, logger, storage, false}).handle()
	}
}
