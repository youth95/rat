package rat

import (
	"fmt"
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
		fmt.Println("startWork")
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
	app.On("/", func(msg interface{}) error {
		headMessage := msg.(*HeadMessage)

		socket := headMessage.Socket
		socket.AddListener("contented", func(msg interface{}) error {
			connected = true
			return nil
		})
		socket.AddListener("message", func(msg interface{}) error {
			ctx := msg.(*MessageContext)
			r = string(ctx.Payload)

			wg.Done()
			return nil
		})
		return nil
	})
	err := app.RunDefault()
	a.Nil(err)
	s = "hello world"
	socket, err := ConnectTimeout("rat://0.0.0.0/", 10*time.Second)
	a.Nil(err)

	_, err = socket.Send([]byte(s), 10*time.Second)
	a.Nil(err)
	wg.Wait()
	a.Equal(s, r)
	a.True(connected)
	a.Nil(err)
	err = app.Close()
	a.Nil(err)
	err = app.ln.Close()
	a.NotNil(err)

}

func TestServer_UseAESM(t *testing.T) {
	a := assert.New(t)
	aesM, err := NewAESMiddleware([]byte("1234567812345679"))
	a.Nil(err)
	var s, r string
	var connected bool
	var wg sync.WaitGroup
	wg.Add(1)
	var app Server
	app.Use(aesM)
	app.On("/", func(msg interface{}) error {
		headMessage := msg.(*HeadMessage)

		socket := headMessage.Socket
		socket.AddListener("contented", func(msg interface{}) error {
			connected = true
			return nil
		})
		socket.AddListener("message", func(msg interface{}) error {
			ctx := msg.(*MessageContext)
			r = string(ctx.Payload)
			wg.Done()
			return nil
		})
		return nil
	})
	err = app.RunDefault()
	a.Nil(err)
	s = "hello world"
	socket, err := ConnectTimeoutWithMiddleware("rat://0.0.0.0/", 10*time.Second, []Middleware{aesM})
	a.Nil(err)
	err = socket.StartWork()
	a.Nil(err)
	_, err = socket.Send([]byte(s), 10*time.Second)
	a.Nil(err)
	wg.Wait()
	a.Equal(s, r)
	a.True(connected)
	a.Nil(err)
	err = app.Close()
	a.Nil(err)
	err = app.ln.Close()
	a.NotNil(err)

}

func TestServer_Request(t *testing.T) {
	a := assert.New(t)
	var s, r string
	var app Server
	app.On("/", func(msg interface{}) error {
		headMessage := msg.(*HeadMessage)
		socket := headMessage.Socket
		socket.AddListener("message", func(msg interface{}) error {
			ctx := msg.(*MessageContext)
			r = string(ctx.Payload)
			_, err := ctx.Send([]byte(r+" back"), ctx.Timeout.Sub(time.Now()))
			if err != nil {
				return err
			}
			return nil
		})
		return nil
	})
	err := app.RunDefault()
	a.Nil(err)
	s = "hello world"
	socket, err := ConnectTimeout("rat://0.0.0.0/", 10*time.Second)
	a.Nil(err)

	res, err := socket.Request([]byte(s))
	a.Nil(err)
	a.Equal(s+" back", string(res.Payload))
	err = app.Close()
	a.Nil(err)
	err = app.ln.Close()
	a.NotNil(err)

}

func TestMessageContext_Reply(t *testing.T) {
	a := assert.New(t)
	var s, r string
	var app Server
	app.On("/", func(msg interface{}) error {
		headMessage := msg.(*HeadMessage)
		socket := headMessage.Socket
		socket.AddListener("message", func(msg interface{}) error {
			ctx := msg.(*MessageContext)
			r = string(ctx.Payload)
			err := ctx.ReplyStringF("%s back", r)
			if err != nil {
				return err
			}
			return nil
		})
		return nil
	})
	err := app.RunDefault()
	a.Nil(err)
	s = "hello world"
	socket, err := ConnectTimeout("rat://0.0.0.0/", 10*time.Second)
	a.Nil(err)

	res, err := socket.RequestString(s)
	a.Nil(err)
	a.Equal(s+" back", string(res.Payload))
	err = app.Close()
	a.Nil(err)
	err = app.ln.Close()
	a.NotNil(err)

}
