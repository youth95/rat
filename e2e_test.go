package rat

import (
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"testing"
	"time"
)

func TestSocket_StartWork(t *testing.T) {
	a := assert.New(t)
	var s, r string
	var connected bool
	var wg sync.WaitGroup
	ln, err := net.Listen("tcp", "0.0.0.0:3321")
	a.Nil(err)
	a.NotEmpty(ln)
	wg.Add(1)
	go func() {
		conn, err := ln.Accept()
		a.Nil(err)
		a.NotEmpty(conn)
		socket := NewSocket(conn)

		socket.Once("contented", func(msg interface{}) error {
			connected = true
			return nil
		})
		socket.Once("message", func(msg interface{}) error {
			ctx := msg.(*MessageContext)
			r = string(ctx.Payload)
			wg.Done()
			return nil
		})

		err = socket.StartWork()
		a.Nil(err)
	}()
	s = "hello world"
	conn, err := net.Dial("tcp", "0.0.0.0:3321")
	socket := NewSocket(conn)
	_, err = socket.Send([]byte(s), 10*time.Second)
	a.Nil(err)
	wg.Wait()
	a.Equal(s, r)
	a.True(connected)
	err = ln.Close()
	a.Nil(err)
	err = conn.Close()
	a.Nil(err)
}

func TestSocket_StartWork2(t *testing.T) {
	a := assert.New(t)
	var wg sync.WaitGroup
	var discontented bool
	ln, err := net.Listen("tcp", "0.0.0.0:3321")
	a.Nil(err)
	a.NotEmpty(ln)
	wg.Add(1)
	go func() {
		conn, err := ln.Accept()
		a.Nil(err)
		a.NotEmpty(conn)
		socket := NewSocket(conn)
		socket.Once("discontented", func(msg interface{}) error {
			discontented = true
			wg.Done()
			return nil
		})
		err = socket.StartWork()
		a.Nil(err)
	}()
	conn, err := net.Dial("tcp", "0.0.0.0:3321")
	a.Nil(err)
	err = conn.Close()
	wg.Wait()
	a.Nil(err)
	a.True(discontented)
	err = ln.Close()
	a.Nil(err)
}

func TestServer_Run(t *testing.T) {
	a := assert.New(t)
	var s, r string
	var connected bool
	var wg sync.WaitGroup
	wg.Add(1)
	var app Server
	app.On("connect", func(msg interface{}) error {
		socket := msg.(*Socket)
		socket.Once("contented", func(msg interface{}) error {
			connected = true
			return nil
		})
		socket.Once("message", func(msg interface{}) error {
			ctx := msg.(*MessageContext)
			r = string(ctx.Payload)
			wg.Done()
			return nil
		})
		return nil
	})
	err := app.Run("0.0.0.0:3321")
	a.Nil(err)
	s = "hello world"
	conn, err := net.Dial("tcp", "0.0.0.0:3321")
	socket := NewSocket(conn)
	_, err = socket.Send([]byte(s), 10*time.Second)
	a.Nil(err)
	wg.Wait()
	a.Equal(s, r)
	a.True(connected)
	a.Nil(err)
	err = conn.Close()
	a.Nil(err)
	err = app.Close()
	a.Nil(err)
}
