package rat

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"time"
)

type Server struct {
	EventEmitter
	started          bool
	ln               net.Listener
	globalMiddleware []Middleware
}

type HeadMessage struct {
	*Socket
	*url.URL
}

func (server *Server) handHeadMessage(socket *Socket, head *Message) error {
	urlInfo, err := url.ParseRequestURI(string(head.Payload))
	if err != nil {
		return err
	}
	ok, err := server.Emit(urlInfo.Path, &HeadMessage{socket, urlInfo}) // dispatch message
	if err != nil {
		_, err = socket.Send([]byte{2}, head.Timeout.Sub(time.Now())) // 响应head包
		return err
	}
	if ok {
		_, err = socket.Send([]byte{0}, head.Timeout.Sub(time.Now())) // 响应head包
	} else {
		_, err = socket.Send([]byte{1}, head.Timeout.Sub(time.Now())) // 响应head包
		return errors.New(fmt.Sprintf("not found path %s", urlInfo.Path))
	}
	return nil
}

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
	server.Once("start", func(msg interface{}) error {
		server.started = true
		return nil
	})
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	server.ln = ln
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				emitError(err)
				continue
			}
			go func() {
				socket := NewSocket(conn)
				socket.Uses(server.globalMiddleware)
				head, err := socket.ReceiveTimeout(10 * time.Second)

				if err != nil {
					emitError(err)
					return
				}
				_, err = server.Emit("connect", &MessageContext{head, socket})
				if err != nil {
					emitError(err)
					return
				}
				err = server.handHeadMessage(socket, head)
				if err != nil {
					emitError(err)
					return
				}
				// 开启工作进程
				err = socket.StartWork()
				if err != nil {
					emitError(err)
				}
			}()
		}
	}()
	return nil
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
