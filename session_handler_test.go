package main

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

type ConnectionMock struct {
	bytes.Buffer
}

func (b *ConnectionMock) Close() error {
	return nil
}

func init() {
	password = ""
}

func TestAuthCommand(t *testing.T) {
	conn := ConnectionMock{}
	_, hook := test.NewNullLogger()
	storage := make(map[string]string)

	handler := sessionHandler{&conn, hook.LastEntry(), storage, false}
	password = "secret"

	err := handler.authCommand([]string{"secret"})
	assert.Nil(t, err)
	assert.True(t, handler.authorized)
	assert.Equal(t, "OK\n", conn.String())

	conn.Reset()
	err = handler.authCommand([]string{})
	assert.Equal(t, "(error) ERR wrong number of arguments for 'auth' command\n", conn.String())

	conn.Reset()
	err = handler.authCommand([]string{"invalid"})
	assert.Equal(t, "(error) ERR invalid password\n", conn.String())
	assert.False(t, handler.authorized)

	conn.Reset()
	password = ""
	err = handler.authCommand([]string{"foo"})
	assert.Equal(t, "(error) ERR Client sent AUTH, but no password is set\n", conn.String())
}

func TestSetCommand(t *testing.T) {
	conn := ConnectionMock{}
	_, hook := test.NewNullLogger()
	storage := make(map[string]string)

	handler := sessionHandler{&conn, hook.LastEntry(), storage, true}

	err := handler.setCommand([]string{"foo", "bar"})
	assert.Nil(t, err)
	assert.Equal(t, "bar", handler.storage["foo"])
	assert.Equal(t, "OK\n", conn.String())

	conn.Reset()
	err = handler.setCommand([]string{"foo", "bar", "cat"})
	assert.Equal(t, "(error) ERR wrong number of arguments for 'set' command\n", conn.String())
}

func TestGetCommand(t *testing.T) {
	conn := ConnectionMock{}
	_, hook := test.NewNullLogger()
	storage := make(map[string]string)

	handler := sessionHandler{&conn, hook.LastEntry(), storage, true}

	err := handler.getCommand([]string{"foo"})
	assert.Equal(t, "(nil)\n", conn.String())
	storage["foo"] = "bar"

	conn.Reset()
	err = handler.getCommand([]string{"foo"})
	assert.Nil(t, err)
	assert.Equal(t, "\"bar\" \n", conn.String())

	conn.Reset()
	err = handler.getCommand([]string{"foo", "cat"})
	assert.Equal(t, "(error) ERR wrong number of arguments for 'get' command\n", conn.String())
}

func TestPingCommand(t *testing.T) {
	conn := ConnectionMock{}
	_, hook := test.NewNullLogger()

	handler := sessionHandler{&conn, hook.LastEntry(), nil, true}

	err := handler.pingCommand([]string{})
	assert.Nil(t, err)
	assert.Equal(t, "PONG\n", conn.String())

	conn.Reset()
	err = handler.pingCommand([]string{"foo", "cat"})
	assert.Equal(t, "(error) ERR wrong number of arguments for 'ping' command\n", conn.String())
}

func TestNoAuth(t *testing.T) {
	conn := ConnectionMock{}
	_, hook := test.NewNullLogger()
	password = "secret"

	handler := sessionHandler{&conn, hook.LastEntry(), nil, false}

	err := handler.setCommand([]string{"foo", "bar"})
	assert.Equal(t, ErrAuth, err)

	err = handler.getCommand([]string{"foo"})
	assert.Equal(t, ErrAuth, err)

	err = handler.pingCommand([]string{})
	assert.Equal(t, ErrAuth, err)
}
