package rat

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestNewSocket(t *testing.T) {
	a := assert.New(t)
	f, err := os.Create("/tmp/a")
	a.Nil(err)
	socket := NewSocket(f)
	a.NotEmpty(socket)
}

func TestSocket_Send(t *testing.T) {
	a := assert.New(t)
	f, err := os.Create("/tmp/a")
	a.Nil(err)
	socket := NewSocket(f)
	msg := []byte("hello world")
	n, err := socket.Send(msg, 1) // 肯定超时
	a.Nil(err)
	a.NotEmpty(n)
	n, err = socket.Send(msg, 60*time.Second)
	a.Nil(err)
	a.NotEmpty(n)
}

func TestSocket_Receive(t *testing.T) {
	a := assert.New(t)
	f, err := os.Open("/tmp/a")
	a.Nil(err)
	socket := NewSocket(f)
	msg, err := socket.Receive()
	a.Nil(err)
	a.Equal(string(msg.Payload), "hello world")
}
