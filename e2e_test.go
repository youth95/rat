package rat

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestServer_Run(t *testing.T) {
	a := assert.New(t)
	var s string
	var connected bool
	var wg sync.WaitGroup
	wg.Add(1)
	var app Server
	app.On(SYSEventConnect, func(msg interface{}) error {
		fmt.Println("SYSEventConnect", msg)
		socket := msg.(*Socket)
		socket.Hand("message", func(payload []byte) []byte {
			fmt.Println(payload)
			return []byte{0}
		})
		return nil
	})
	err := app.RunDefault()
	a.Nil(err)
	s = "hello world"
	socket, err := Connect(ConnectOption{
		Uri: "rat://0.0.0.0/",
	})
	a.Nil(err)
	mr, err := socket.RequestString("message", s)
	a.Nil(err)
	wg.Wait()
	a.Equal(s, string(mr.Payload))
	a.True(connected)
	a.Nil(err)
	err = app.Close()
	a.Nil(err)
	err = app.ln.Close()
	a.NotNil(err)

}

//func TestServer_UseAESM(t *testing.T) {
//	a := assert.New(t)
//	aesM, err := NewAESMiddleware([]byte("1234567812345679"))
//	a.Nil(err)
//	var s, r string
//	var connected bool
//	var wg sync.WaitGroup
//	wg.Add(1)
//	var app Server
//	app.Use(aesM)
//	app.On("/", func(msg interface{}) error {
//		headMessage := msg.(*HeadMessage)
//
//		socket := headMessage.Socket
//		socket.AddListener("contented", func(msg interface{}) error {
//			connected = true
//			return nil
//		})
//		socket.AddListener("message", func(msg interface{}) error {
//			ctx := msg.(*MessageContext)
//			r = string(ctx.Payload)
//			wg.Done()
//			return nil
//		})
//		return nil
//	})
//	err = app.RunDefault()
//	a.Nil(err)
//	s = "hello world"
//	socket, err := ConnectTimeoutWithMiddleware("rat://0.0.0.0/", 10*time.Second, []Middleware{aesM})
//	a.Nil(err)
//	err = socket.StartWork()
//	a.Nil(err)
//	_, err = socket.Send([]byte(s), 10*time.Second)
//	a.Nil(err)
//	wg.Wait()
//	a.Equal(s, r)
//	a.True(connected)
//	a.Nil(err)
//	err = app.Close()
//	a.Nil(err)
//	err = app.ln.Close()
//	a.NotNil(err)
//
//}
//
//func TestServer_Request(t *testing.T) {
//	a := assert.New(t)
//	var s, r string
//	var app Server
//	app.On("/", func(msg interface{}) error {
//		headMessage := msg.(*HeadMessage)
//		socket := headMessage.Socket
//		socket.AddListener("message", func(msg interface{}) error {
//			ctx := msg.(*MessageContext)
//			r = string(ctx.Payload)
//			_, err := ctx.Send([]byte(r+" back"), ctx.Timeout.Sub(time.Now()))
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//		return nil
//	})
//	err := app.RunDefault()
//	a.Nil(err)
//	s = "hello world"
//	socket, err := ConnectTimeout("rat://0.0.0.0/", 10*time.Second)
//	a.Nil(err)
//
//	res, err := socket.Request([]byte(s))
//	a.Nil(err)
//	a.Equal(s+" back", string(res.Payload))
//	err = app.Close()
//	a.Nil(err)
//	err = app.ln.Close()
//	a.NotNil(err)
//
//}
//
//func TestMessageContext_Reply(t *testing.T) {
//	a := assert.New(t)
//	var s, r string
//	var app Server
//	app.On("/", func(msg interface{}) error {
//		headMessage := msg.(*HeadMessage)
//		socket := headMessage.Socket
//		socket.AddListener("message", func(msg interface{}) error {
//			ctx := msg.(*MessageContext)
//			r = string(ctx.Payload)
//			err := ctx.ReplyStringF("%s back", r)
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//		return nil
//	})
//	err := app.RunDefault()
//	a.Nil(err)
//	s = "hello world"
//	socket, err := ConnectTimeout("rat://0.0.0.0/", 10*time.Second)
//	a.Nil(err)
//
//	res, err := socket.RequestString(s)
//	a.Nil(err)
//	a.Equal(s+" back", string(res.Payload))
//	err = app.Close()
//	a.Nil(err)
//	err = app.ln.Close()
//	a.NotNil(err)
//
//}
