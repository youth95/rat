package rat

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
)

var SYSEventConnect = "connect"

type Server struct {
	EventEmitter
	started          bool
	ln               net.Listener
	globalMiddleware []Middleware
	EnterPoint       map[string]EnterHandler
}

type HeadMessage struct {
	*Socket
	*url.URL
}

type MessageContext struct {
	*Message
	*Socket
}

type EnterHandler func(path string, url string, socket *Socket)

func (server *Server) Run(addr string) error {
	if server.started == true {
		return errors.New("server already started")
	}
	server.started = true
	emitError := func(err error) {
		has, err := server.Emit("error", err)
		if err != nil {
			if has == false {
				log.Printf("error: %s", err.Error())
			}
			panic(err)
			return
		}

	}
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	ln, err := net.ListenTCP("tcp", rAddr)
	if err != nil {
		return err
	}
	server.ln = ln
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			emitError(err)
			continue
		}
		socket := NewSocket(NewTcpConnFull(conn), server.globalMiddleware)
		// client must Request a uri
		msg, err := socket.ReceiveMessage(DefaultRequestTimeOut)
		if err != nil {
			emitError(err)
			_ = conn.Close()
			continue
		}
		urlStr := string(msg.Payload)
		urlInfo, err := url.ParseRequestURI(urlStr)
		if err != nil {
			emitError(err)
			_ = conn.Close()
			continue
		}
		if server.EnterPoint[urlInfo.Path] != nil {
			err = socket.SendResponse([]byte{0}, DefaultRequestTimeOut) // 响应head包
			server.EnterPoint[urlInfo.Path](urlInfo.Path, urlStr, socket)
		} else {
			err = socket.SendResponse([]byte{1}, DefaultRequestTimeOut) // 响应head包
			return errors.New(fmt.Sprintf("not found path %s", urlInfo.Path))
		}
	}
}

func (server *Server) RunDefault() error {
	return server.Run(fmt.Sprintf("0.0.0.0:%d", DefaultServerPort))
}

// 关闭连接
func (server *Server) Close() error {
	if server.started {
		err := server.ln.Close()
		if err != nil {
			return err
		}
		server.started = false
		return nil
	} else {
		return errors.New("server not started yet")
	}
}

func (server *Server) Use(m Middleware) {
	server.globalMiddleware = append(server.globalMiddleware, m)
}
