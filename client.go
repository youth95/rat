package rat

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	*Socket
	entryPoint string
}

type ConnectOption struct {
	Uri        string
	Timeout    time.Duration
	Middleware []Middleware
	AutoReConnect bool
}

func NewClient(opt ConnectOption) (*Client, error) {
	timeout := opt.Timeout
	if timeout == 0 {
		timeout = DefaultConnectTimeOut
	}
	ms := opt.Middleware
	if ms == nil {
		ms = []Middleware{}
	}
	uri := opt.Uri
	if uri == "" {
		return nil, ErrorUriIsRequired
	}
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
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host, portStr))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}
	tcpConnFull := NewTcpConnFull(conn)
	socket := NewSocket(tcpConnFull, ms)
	return &Client{Socket: socket, entryPoint: uri}, nil
}

func (client *Client) Connect() error {
	err := client.SendRequest([]byte(client.entryPoint), DefaultConnectTimeOut)
	if err != nil {
		return err
	}
	res, err := client.ReceiveMessage(DefaultConnectTimeOut)
	if err != nil {
		return err
	}
	switch res.Payload[0] {
	case 1:
		return errors.New(fmt.Sprintf("not found entry point %s", client.entryPoint))
	case 2:
		return errors.New("server error")
	}
	return nil
}
