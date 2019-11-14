package rat

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"
)

const (
	DefaultServerPort = 3399
)

type Message struct {
	Timeout *time.Time
	Length  int
	Payload []byte
}

type Socket struct {
	EventEmitter
	conn       io.ReadWriter
	beginStart bool
	stop       bool
	middleware []Middleware
}

type MessageContext struct {
	*Message
	*Socket
}

// 响应消息
func (messageContext *MessageContext) Reply(payload []byte) error {
	_, err := messageContext.Send(payload, messageContext.Timeout.Sub(time.Now()))
	return err
}

func (messageContext *MessageContext) ReplyString(payload string) error {
	return messageContext.Reply([]byte(payload))
}

func (messageContext *MessageContext) ReplyStringF(format string, a ...interface{}) error {
	return messageContext.ReplyString(fmt.Sprintf(format, a...))
}

func NewSocket(conn io.ReadWriter) *Socket {
	return &Socket{EventEmitter{}, conn, false, false, []Middleware{}}
}

func ConnectTimeout(uri string, timeout time.Duration) (*Socket, error) {
	return ConnectTimeoutWithMiddleware(uri, timeout, nil)
}

// ConnectTimeoutWithMiddleware
func ConnectTimeoutWithMiddleware(uri string, timeout time.Duration, ms []Middleware) (*Socket, error) {
	connectInfo, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}
	if connectInfo.Scheme != "rat" {
		return nil, errors.New("connect uri.Scheme must be 'rat'")
	}
	portStr := connectInfo.Port()
	if portStr == "" {
		portStr = strconv.FormatInt(DefaultServerPort, 10)
	}
	host := connectInfo.Hostname()

	addr := fmt.Sprintf("%s:%s", host, portStr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	socket := NewSocket(conn)
	if ms != nil {
		socket.middleware = ms
	}
	// 握手
	res, err := socket.RequestTimeout([]byte(uri), timeout)
	if err != nil {
		return nil, err
	}
	switch res.Payload[0] {
	case 1:
		return nil, errors.New(fmt.Sprintf("not found path %s", connectInfo.Path))
	case 2:
		return nil, errors.New("server error")
	}
	return socket, nil
}

// Receive 读取一个没有超时的消息 若limit时间之后仍然没读到,则返回一个错误
func (socket *Socket) ReceiveTimeout(limit time.Duration) (*Message, error) {
	errCH := make(chan error)
	msgCH := make(chan *Message)

	go func() {
		start := time.Now()
		reader := bufio.NewReader(socket.conn)
		for {
			now := time.Now()
			if limit > 0 && start.Add(limit).Before(now) {
				errCH <- errors.New("socket receive timeout")
				return
			}
			timeout, l32, err := ReadFixedHeader(reader)
			l := int(l32)
			if err != nil {
				errCH <- err
				return
			}

			if now.After(*timeout) && !now.Equal(*timeout) { // 如果当前时间在消息时限之后 且 当前时间不等于限时时间则说明该消息已经过时需要跳过
				for l > 0 {
					p := make([]byte, l)
					n, err := reader.Read(p)
					if err != nil {
						errCH <- err
						return
					}
					l -= n
				}
				continue // read next message
			} else {
				var payload []byte
				payloadBuf := make([]byte, l)
				for len(payload) != l {
					n, err := reader.Read(payloadBuf)
					if err != nil {
						errCH <- err
					}
					if n+len(payload) >= l {
						payload = append(payload, payloadBuf[:n]...)
						payload, err = socket.middlewareReadHandData(payload)
						if err != nil {
							errCH <- err
						}
						msgCH <- &Message{timeout, l, payload,}
						return
					} else {
						payload = append(payload, payloadBuf[:n]...)
					}
					payloadBuf = make([]byte, l-(n+len(payload))) // 防止读多了
				}
			}
		}
	}()
	var msg *Message
	var err error
	select {
	case <-time.Tick(limit):
		return nil, errors.New("socket receive timeout")
	case msg = <-msgCH:
		return msg, nil
	case err = <-errCH:
		return nil, err
	}
}

// Receive 读取一个没有超时的消息 无限等待
func (socket *Socket) Receive() (*Message, error) {
	return socket.ReceiveTimeout(-1)
}

// Send 发送一个消息
func (socket *Socket) Send(payload []byte, timeout time.Duration) (int64, error) {
	w := socket.conn
	payload, err := socket.middlewareWriteHandData(payload)
	if err != nil {
		return 0, err
	}
	tl, err := WriteFixedHeaderFromMessage(w, timeout, payload)
	if err != nil {
		return 0, err
	}
	l := len(payload)
	for l != 0 {
		n, err := w.Write(payload)
		if err != nil {
			return 0, err
		}
		l -= n
		payload = payload[n:]
	}
	return tl, nil
}

func (socket *Socket) SendString(payload string, timeout time.Duration) (int64, error) {
	return socket.SendString(payload, timeout)
}

func (socket *Socket) SendStringF(timeout time.Duration, format string, a ...interface{}) (int64, error) {
	return socket.SendString(fmt.Sprintf(format, a...), timeout)
}

// RequestTimeout 请求一个消息并等待响应,可传入超时时间
func (socket *Socket) RequestTimeout(payload []byte, limit time.Duration) (*Message, error) {
	msgCH := make(chan *Message)
	errCH := make(chan error)
	_, err := socket.Send(payload, limit)
	if err != nil {
		return nil, err
	}
	go func() {
		msg, err := socket.ReceiveTimeout(limit)
		if err != nil {

		}
		msgCH <- msg
	}()
	var msg *Message
	select {
	case msg = <-msgCH:
		return msg, nil
	case err = <-errCH:
		return nil, err
	case <-time.Tick(limit):
		return nil, errors.New("request timeout")
	}
}

// Request 请求一个消息并等待响应,默认超时10s
func (socket *Socket) Request(payload []byte) (*Message, error) {
	return socket.RequestTimeout(payload, 10*time.Second)
}

func (socket *Socket) RequestString(payload string) (*Message, error) {
	return socket.Request([]byte(payload))
}

func (socket *Socket) RequestStringF(format string, a ...interface{}) (*Message, error) {
	return socket.RequestString(fmt.Sprintf(format, a...))
}

// StartWork 启动监听
func (socket *Socket) StartWork() error {

	if socket.beginStart {
		return errors.New("socket already work started")
	} else {
		socket.beginStart = true
		socket.stop = false
	}
	_, err := socket.Emit("contented", socket)
	if err != nil {
		return err
	}
	go func() {
		for {
			if socket.stop {
				break
			}
			msg, err := socket.Receive()
			if err != nil {
				if err != io.EOF {
					_, _ = socket.Emit("error", err)
				} else {
					_, _ = socket.Emit("discontented", socket)
					break // 退出
				}
			} else {
				_, _ = socket.Emit("message", &MessageContext{msg, socket})
			}
		}
	}()
	return nil
}

// StopWork 停止监听
func (socket *Socket) StopWork() {
	socket.stop = true
	socket.beginStart = false
}

func (socket *Socket) GetConn() io.ReadWriter {
	return socket.conn
}
