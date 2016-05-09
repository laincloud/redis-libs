package network

import (
	"errors"
	"net"
	"time"
)

var ErrConnNil = errors.New("Connector is Nil")

type Conn struct {
	co  *connectOption
	cn  net.Conn
	err error
}

func NewConnect(cn net.Conn, co *connectOption) *Conn {
	c := &Conn{
		co: co,
		cn: cn,
	}
	return c
}

func (c *Conn) Write(msg string) error {
	if c.cn == nil {
		return ErrConnNil
	}
	c.cn.SetWriteDeadline(time.Now().Add(c.co.wrteTimeOutSec))
	_, c.err = c.cn.Write([]byte(msg))
	return c.err
}

func (c *Conn) Read(b []byte) (int, error) {
	if c.cn == nil {
		return 0, ErrConnNil
	}
	c.cn.SetWriteDeadline(time.Now().Add(c.co.wrteTimeOutSec))
	n, err := c.cn.Read(b)
	c.err = err
	return n, err
}

func (c *Conn) ReadAll() (string, error) {
	if c.cn == nil {
		return "", ErrConnNil
	}

	c.cn.SetReadDeadline(time.Now().Add(c.co.readTimeOutSec))
	res := ""
	bufferSize := c.co.bufferSize
	buffer := make([]byte, bufferSize)
	for {
		n, err := c.Read(buffer)
		if err != nil {
			c.err = err
			return res, err
		}
		res += string(buffer[:n])
		if n < bufferSize {
			break
		}
	}
	return res, nil
}

func (c *Conn) Conn() net.Conn {
	return c.cn
}
func (c *Conn) Close() error {
	if c.cn == nil {
		return ErrConnNil
	}
	return c.cn.Close()
}
