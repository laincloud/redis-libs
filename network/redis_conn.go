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

func NewRedisConn(conn net.Conn) (*RedisConn, error) {
	cn, err := NewConn(conn)
	if err != nil {
		return nil, err
	}
	r := &RedisConn{Conn: cn}
	r.RedisReader = NewRedisReader(bufio.NewReader(r))
	return r, nil
}

func (r *RedisConn) ReadAll() ([]byte, error) {
	return r.ReadObject()
}
