package network

import (
	"errors"
	"net"
	"syscall"
	"time"
)

var ErrConnNil = errors.New("Connector is Nil")

type IConn interface {
	Write(msg []byte) error
	Read(b []byte) (int, error)
	ReadAll() ([]byte, error)
	Close() error
}

type Conn struct {
	conn net.Conn
	cnop *ConnectOption
	err  error
}

func NewConnect(conn net.Conn, co *ConnectOption) *Conn {
	if conn == nil {
		return nil
	}
	c := &Conn{conn: conn, cnop: co}
	return c
}

func (c *Conn) GetConn() net.Conn {
	return c.conn
}

func (c *Conn) Write(b []byte) error {
	if c == nil {
		return ErrConnNil
	}
	c.conn.SetWriteDeadline(time.Now().Add(c.cnop.wrteTimeOutSec))
	size := len(b)
	from := 0
	for {
		n, err := c.conn.Write(b[from:])
		if err != nil {
			if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			c.err = err
			break
		}
		from += n
		if from == size {
			break
		}
	}
	return c.err
}

func (c *Conn) Read(b []byte) (int, error) {
	if c == nil {
		return 0, ErrConnNil
	}
	c.conn.SetReadDeadline(time.Now().Add(c.cnop.wrteTimeOutSec))
	n, err := c.conn.Read(b)
	c.err = err
	return n, err
}

func (c *Conn) ReadAll() ([]byte, error) {
	if c == nil {
		return nil, ErrConnNil
	}
	c.conn.SetReadDeadline(time.Now().Add(c.cnop.readTimeOutSec))
	res := make([]byte, 0, 0)
	bufferSize := c.cnop.bufferSize
	buffer := make([]byte, bufferSize)
	for {
		n, err := c.Read(buffer)
		if err != nil {
			c.err = err
			return res, err
		}
		res = append(res, buffer[:n]...)
		if n < bufferSize {
			break
		}
	}
	return res, nil
}

func (c *Conn) Close() error {
	if c == nil {
		return ErrConnNil
	}
	return c.conn.Close()
}
