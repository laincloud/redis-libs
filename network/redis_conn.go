package network

import (
	"bufio"
	"errors"
	"github.com/mijia/sweb/log"
	"net"
	"strconv"
	"time"
)

const (
	SIMPLE_STRING = '+'
	BULK_STRING   = '$'
	INTEGER       = ':'
	ARRAY         = '*'
	ERROR         = '-'

	SYM_CRLF = "\r\n"
)

var (
	BAD_ELEMENT_ERR = errors.New("-Error Bad Element\r\n")
	NILL_VAL_ERR    = errors.New("-Error Nil Connection\r\n")
)

type RedisConn struct {
	*Conn
	br *bufio.Reader
}

func NewRedisConn(conn net.Conn, co *ConnectOption) (*RedisConn, error) {
	if conn == nil {
		return nil, NILL_VAL_ERR
	}
	r := &RedisConn{Conn: NewConnect(conn, co)}
	r.br = bufio.NewReader(r)
	return r, nil
}

func (r *RedisConn) ReadAll() ([]byte, error) {
	return r.readObject()
}

func (r *RedisConn) readObject() ([]byte, error) {
	r.conn.SetReadDeadline(time.Now().Add(r.cnop.readTimeOutSec))
	line, err := r.readLine()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	switch line[0] {
	case SIMPLE_STRING, INTEGER, ERROR:
		return line, nil
	case BULK_STRING:
		return r.readBulkString(line)
	case ARRAY:
		return r.readArray(line)
	default:
		return nil, BAD_ELEMENT_ERR
	}
}

func (r *RedisConn) readLine() ([]byte, error) {
	p, err := r.br.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *RedisConn) getCount(line []byte) (int, error) {
	if !(line[len(line)-1] == '\n' && line[len(line)-2] == '\r') {
		return 0, BAD_ELEMENT_ERR
	}
	return strconv.Atoi(string(line[1 : len(line)-2]))
}

func (r *RedisConn) readBulkString(line []byte) ([]byte, error) {
	size, err := r.getCount(line)
	if err != nil {
		return nil, err
	}
	res := make([]byte, 0, 0)
	res = append(res, line...)
	if size == -1 {
		return res, nil
	}
	if size < 0 {
		return nil, BAD_ELEMENT_ERR
	}
	size = size + 2
	buffer, err := r.br.Peek(size)
	r.br.Discard(size)
	res = append(res, buffer...)
	return res, err
}

func (r *RedisConn) readArray(line []byte) ([]byte, error) {
	// Get number of array elements.
	count, err := r.getCount(line)
	if err != nil {
		return nil, err
	}
	// Read `count` number of RESP objects in the array.
	buf := make([]byte, 0, 0)
	buf = append(buf, line...)
	for i := 0; i < count; i++ {
		sub, err := r.readObject()
		if err != nil {
			return nil, err
		}
		buf = append(buf, sub...)
	}
	return buf, nil
}
