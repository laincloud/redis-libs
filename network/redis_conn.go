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
	cn, err := NewConnect(conn, co)
	if err != nil {
		return nil, err
	}
	r := &RedisConn{Conn: cn}
	r.br = bufio.NewReader(r)
	return r, nil
}

func (r *RedisConn) ReadAll() ([]byte, error) {
	res := make([]byte, 0)
	err := r.readObject(&res)
	return res, err
}

func (r *RedisConn) readObject(res *[]byte) error {
	r.conn.SetReadDeadline(time.Now().Add(r.cnop.readTimeOutSec))
	line, err := r.readLine()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	*res = append(*res, []byte(line)...)
	switch line[0] {
	case SIMPLE_STRING, INTEGER, ERROR:
		return nil
	case BULK_STRING:
		return r.readBulkString(line, res)
	case ARRAY:
		return r.readArray(line, res)
	default:
		return BAD_ELEMENT_ERR
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

func (r *RedisConn) readBulkString(line []byte, res *[]byte) error {
	size, err := r.getCount(line)
	if err != nil {
		return err
	}
	if size == -1 {
		return nil
	}
	if size < 0 {
		return BAD_ELEMENT_ERR
	}
	size = size + 2
	buffer, err := r.br.Peek(size)
	r.br.Discard(size)
	*res = append(*res, buffer...)
	return err
}

func (r *RedisConn) readArray(line []byte, res *[]byte) error {
	// Get number of array elements.
	count, err := r.getCount(line)
	if err != nil {
		return err
	}
	// Read `count` number of RESP objects in the array.
	for i := 0; i < count; i++ {
		err := r.readObject(res)
		if err != nil {
			return err
		}
	}
	return nil
}
