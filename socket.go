package rat

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type Message struct {
	*FixedHeader
	Payload []byte
}

type RequestMessage struct {
	Length  int
	Event   string
	Payload []byte
}

type ResponseMessage struct {
	Length  int
	Payload []byte
}

const (
	DefaultConnectTimeOut   = 10 * time.Second
	DefaultRequestTimeOut   = 10 * time.Second
	DefaultHBTimeOut        = 3 * time.Second
	DefaultServerPort       = 3399
	SocketEventError        = "error"
	SocketEventRequest      = "request"
	SocketEventResponse     = "response"
	SocketEventHB           = "hb"
	SocketEventDisconnected = "disconnected"
)

var ErrorUriIsRequired = errors.New("uri is required")

type Socket struct {
	EventEmitter
	conn       *TcpConnFull
	Conn       *net.TCPConn
	middleware []Middleware
	ms         *sync.Mutex
	handlerMap sync.Map
	closed     bool
}

func NewSocket(conn *TcpConnFull, ms []Middleware) *Socket {
	socket := &Socket{EventEmitter{}, conn, conn.TCPConn, ms, &sync.Mutex{}, sync.Map{}, false}
	return socket
}

// SendRequest 发送一个请求
func (socket *Socket) SendRequest(payload []byte, limit time.Duration) error {
	var fh FixedHeader
	fh.MT = MTRequest
	fh.ML = int32(len(payload))
	err := socket.conn.WriteFull(fh.Bytes(), limit)
	if err != nil {
		return err
	}
	payload, err = socket.middlewareWriteHandData(payload)
	if err != nil {
		return err
	}
	err = socket.conn.WriteFull(payload, limit)
	if err != nil {
		return err
	}
	return nil
}

// SendResponse 发送一个响应
func (socket *Socket) SendResponse(payload []byte, limit time.Duration) error {
	var fh FixedHeader
	fh.MT = MTResponse
	fh.ML = int32(len(payload))
	err := socket.conn.WriteFull(fh.Bytes(), limit)
	if err != nil {
		return err
	}
	payload, err = socket.middlewareWriteHandData(payload)
	if err != nil {
		return err
	}
	err = socket.conn.WriteFull(payload, limit)
	if err != nil {
		return err
	}
	return nil
}

// SendHB 发送心跳
func (socket *Socket) SendHB(limit time.Duration) error {
	var fh FixedHeader
	fh.MT = MTHB
	fh.ML = 0
	err := socket.conn.WriteFull(fh.Bytes(), limit)
	if err != nil {
		return err
	}
	return nil
}

// Request 请求
func (socket *Socket) Request(payload []byte, limit time.Duration) ([]byte, error) {
	payloadCH := make(chan []byte)
	socket.Once(SocketEventResponse, func(msg interface{}) error {
		payloadCH <- msg.([]byte)
		return nil
	})
	err := socket.SendRequest(payload, limit)
	if err != nil {
		return nil, err
	}

	return <-payloadCH, nil
}

// HB 只有开启HB的时候 上层应用才能检测到socket连接断开
func (socket *Socket) HB() {
	for {
		time.Sleep(DefaultHBTimeOut) // TODO 心跳间隔应该由另外的变量表示
		err := socket.SendHB(DefaultHBTimeOut)
		if err != nil {
			_, _ = socket.Emit(SocketEventDisconnected, nil)
			socket.closed = true
			err := socket.conn.Close()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("hb go func exit")

			return
		}
	}
}

// Loop 只有开启Loop的时候 上层应用才能使用Request方法
func (socket *Socket) Loop(limit time.Duration) {
	for {
		msg, err := socket.ReceiveMessage(limit)
		if err != nil {
			_, _ = socket.Emit(SocketEventError, err)
			if socket.closed {
				fmt.Println("loop fun exit")
				return
			} else {
				continue
			}
		}
		switch msg.MT {
		case MTRequest:
			_, _ = socket.Emit(SocketEventRequest, msg.Payload)
			break
		case MTResponse:
			_, _ = socket.Emit(SocketEventResponse, msg.Payload)
			break
		case MTHB:
			_, _ = socket.Emit(SocketEventHB, nil)
			break
		default:
			_, _ = socket.Emit(SocketEventError, errors.New("unknown MT"))
			return
		}
	}
}

func (socket *Socket) ReceiveMessage(limit time.Duration) (*Message, error) {
	var msg Message
	buf := make([]byte, FixedHeaderSize)
	err := socket.conn.ReadFull(buf, limit)
	if err != nil {
		return nil, err
	}
	fh := ParseFixedHeader(buf)
	msg.FixedHeader = fh
	if fh.ML == 0 {
		msg.Payload = []byte{}
		return &msg, nil
	}
	buf = make([]byte, fh.ML)
	err = socket.conn.ReadFull(buf, limit)
	if err != nil {
		return nil, err
	}
	buf, err = socket.middlewareReadHandData(buf)
	if err != nil {
		return nil, err
	}
	msg.Payload = buf
	return &msg, nil
}

func (socket *Socket) CloseLoop() {
	socket.closed = true
}
