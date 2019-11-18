package rat

import (
	"io"
	"net"
	"sync"
	"time"
)

type TcpConnFull struct {
	*net.TCPConn
	r *sync.Mutex
	w *sync.Mutex
}

func NewTcpConnFull(conn *net.TCPConn) *TcpConnFull {
	return &TcpConnFull{conn, &sync.Mutex{}, &sync.Mutex{}}
}

// ReadFull only return nil or io.EOF
func (conn *TcpConnFull) ReadFull(b []byte, limit time.Duration) error {
	conn.r.Lock()
	defer conn.r.Unlock()

	var result []byte

	l := len(b)
	for l > 0 {
		buf := make([]byte, l)
		err := conn.SetReadDeadline(time.Now().Add(limit))
		if err != nil {
			return io.EOF
		}
		n, err := conn.Read(buf)
		if err != nil {
			return io.EOF
		}
		l -= n
		result = append(result, buf[:n]...)
	}
	for i, v := range result {
		b[i] = v
	}
	err := conn.SetReadDeadline(time.Time{})
	if err != nil {
		return io.EOF
	}
	//b = result
	//fmt.Println("ReadFull", b)
	return nil
}

// WriteFull only return nil or io.EOF
func (conn *TcpConnFull) WriteFull(p []byte, limit time.Duration) error {
	conn.w.Lock()
	defer conn.w.Unlock()
	//fmt.Println("WriteFull", p)

	l := len(p)
	for l > 0 {
		err := conn.SetWriteDeadline(time.Now().Add(limit))
		if err != nil {
			return io.EOF
		}
		n, err := conn.Write(p)
		if err != nil {
			return io.EOF
		}
		l -= n
	}
	err := conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return io.EOF
	}
	return nil
}
