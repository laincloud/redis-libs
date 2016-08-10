package network

import (
	"bufio"
	"errors"
	"net"
	// "time"
)

var (
	BAD_ELEMENT_ERR = errors.New("-Error Bad Element\r\n")
	NILL_VAL_ERR    = errors.New("-Error Nil Connection\r\n")
)

type RedisConn struct {
	*Conn
	*RedisReader
}

func NewRedisConn(conn net.Conn, co *ConnectOption) (*RedisConn, error) {
	cn, err := NewConnect(conn, co)
	if err != nil {
		return nil, err
	}
	r := &RedisConn{Conn: cn}
	r.RedisReader = NewRedisReader(bufio.NewReader(r))
	return r, nil
}

func (r *RedisConn) ReadAll() ([]byte, error) {
	// r.conn.SetReadDeadline(time.Now().Add(r.cnop.readTimeOutSec))
	return r.ReadObject()
}
