package rat

import (
	"bufio"
	"errors"
	"io"
	"time"
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
}

type MessageContext struct {
	Message
	Socket
}

func NewSocket(conn io.ReadWriter) *Socket {
	return &Socket{EventEmitter{}, conn, false}
}

// Receive 读取一个没有超时的消息
func (socket *Socket) Receive() (*Message, error) {
	reader := bufio.NewReader(socket.conn)
	for {
		now := time.Now()
		timeout, l32, err := ReadFixedHeader(reader)
		l := int(l32)
		if err != nil {
			return nil, err
		}

		if now.After(*timeout) && !now.Equal(*timeout) { // 如果当前时间在消息时限之后 且 当前时间不等于限时时间则说明该消息已经过时需要跳过
			for l > 0 {
				p := make([]byte, l)
				n, err := reader.Read(p)
				if err != nil {
					return nil, err
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
					return nil, err
				}
				if n+len(payload) >= l {
					payload = append(payload, payloadBuf[:n]...)
					return &Message{timeout, l, payload,}, nil
				} else {
					payload = append(payload, payloadBuf[:n]...)
				}
				payloadBuf = make([]byte, l-(n+len(payload))) // 防止读多了
			}
		}
	}
}

// Send 发送一个消息
func (socket *Socket) Send(payload []byte, timeout time.Duration) (int64, error) {
	w := socket.conn
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

// StartWork 启动监听
func (socket *Socket) StartWork() error {
	if socket.beginStart {
		return errors.New("socket already work started")
	} else {
		socket.beginStart = true
	}
	_, err := socket.Emit("contented", socket)
	if err != nil {
		return err
	}
	go func() {
		for {
			msg, err := socket.Receive()
			if err != nil {
				if err != io.EOF {
					_, _ = socket.Emit("error", err)
				}else {
					_, _ = socket.Emit("discontented", socket)
					break	// 退出
				}
			} else {
				_, _ = socket.Emit("message", &MessageContext{*msg, *socket})
			}
		}
	}()
	return nil
}
