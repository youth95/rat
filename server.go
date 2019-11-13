package rat

import (
	"errors"
	"log"
	"net"
)

type Server struct {
	EventEmitter
	started bool
	ln      net.Listener
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
			socket := NewSocket(conn)
			_, err = server.Emit("connect", socket)
			if err != nil {
				emitError(err)
				continue
			}
			err = socket.StartWork()
			if err != nil {
				emitError(err)
			}
		}
	}()
	return nil
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
