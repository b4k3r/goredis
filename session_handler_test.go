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
