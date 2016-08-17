package network

import (
	"errors"
	"net"
	"syscall"
	"time"
)

var ErrConnNil = errors.New("Connector is Nil")

type IConn interface {
	net.Conn
	ReadAll() ([]byte, error)
	ShouldBeClosed() bool
}

type Conn struct {
	net.Conn
	err error
}

func NewConn(conn net.Conn) (*Conn, error) {
	if conn == nil {
		return nil, ErrConnNil
	}
	c := &Conn{Conn: conn}
	return c, nil
}

func (c *Conn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	c.err = err
	return n, c.err
}

func (c *Conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.err = err
	return n, c.err
}

func (c *Conn) ReadAll() ([]byte, error) {
	return nil, syscall.ENOSYS
}

func (c *Conn) ShouldBeClosed() bool {
	return c.err != nil
}

type TimeoutConn struct {
	*Conn
	cnop *ConnectOption
}

func NewTimeoutConn(conn net.Conn, cnop *ConnectOption) (*TimeoutConn, error) {
	cn, err := NewConn(conn)
	if err != nil {
		return nil, err
	}
	c := &TimeoutConn{Conn: cn, cnop: cnop}
	return c, nil
}

func (c *TimeoutConn) Write(b []byte) (int, error) {
	c.SetWriteDeadline(time.Now().Add(c.cnop.wrteTimeOutSec))
	return c.Conn.Write(b)
}

func (c *TimeoutConn) Read(b []byte) (int, error) {
	c.SetReadDeadline(time.Now().Add(c.cnop.wrteTimeOutSec))
	return c.Conn.Read(b)
}

func (c *TimeoutConn) ReadAll() ([]byte, error) {
	c.SetReadDeadline(time.Now().Add(c.cnop.readTimeOutSec))
	return c.ReadAll()
}
