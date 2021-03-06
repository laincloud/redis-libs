package redislibs

import (
	"errors"
	"strconv"
	"strings"
)

var (
	arrayPrefixSlice      = []byte{'*'}
	bulkStringPrefixSlice = []byte{'$'}
	lineEndingSlice       = []byte{'\r', '\n'}
)

const (
	SIMPLE_STRING = '+'
	BULK_STRING   = '$'
	INTEGER       = ':'
	ARRAY         = '*'
	ERROR         = '-'
)

const (
	SYM_ERROR  = "-"
	SYM_STAR   = "*"
	SYM_DOLLAR = "$"
	SYM_CRLF   = "\r\n"
	SYM_EMPTY  = ""
	SYM_LF     = "\n"
	SYM_MINUS  = "-"

	NETADDRSPLITSMB = ":"
	ROLESPLITSMB    = ","

	SLOT_COUNT = 16384

	WIND_MIGRATE    = 100
	TIMEOUT_MIGRATE = 1000

	STATUS_OK  = "OK"
	STATUS_ERR = "ERR"

	ROLE_STATUS_ERR       = "-ERR"
	ROLE_STATUS_CONNECT   = "connect"
	ROLE_STATUS_SYNC      = "sync"
	ROLE_STATUS_CONNECTED = "connected"

	REDIS_DB_NO = 0

	ROLE_MASTER   = "master"
	ROLE_SENTINEL = "sentinel"
	ROLE_SLAVE    = "slave"
	ROLE_MYSELF   = "myself"
	ROLE_FAIL     = "fail"
	ROLE_NOADDR   = "noaddr"

	INFO_TIME_OUT      = 15
	INFO_CLUSTER_STATE = "cluster_state"

	DEFAULT_ADDRESS = "127.0.0.1"
)

var (
	COMMAND_CLUSTER_NODES = Pack_command("CLUSTER", "NODES")
	COMMAND_INFO          = Pack_command("INFO")
	COMMAND_PING          = Pack_command("PING")
)

func Pack_command(args ...string) string {
	buf := ""
	buf += SYM_STAR + strconv.Itoa(len(args)) + SYM_CRLF
	for _, v := range args {
		buf += SYM_DOLLAR + strconv.Itoa(len(v)) + SYM_CRLF + v + SYM_CRLF
	}
	return buf
}

func (t *Talker) TalkRaw(commands string) (string, error) {
	_, err := t.Write([]byte(commands))
	if err != nil {
		return "", nil
	}
	res, err := t.ReadAll()
	return string(res), err
}

func (t *Talker) TalkForObject(commands string) (interface{}, error) {
	_, err := t.Write([]byte(commands))
	if err != nil {
		return "", nil
	}
	return t.readObject()
}

func (t *Talker) readObject() (interface{}, error) {
	line, err := t.readLine()
	if err != nil {
		return nil, err
	}
	switch line[0] {
	case SIMPLE_STRING, INTEGER, ERROR:
		return string(line[1:]), nil
	case BULK_STRING:
		if size, err := t.getCount([]byte(line)); err != nil {
			return nil, err
		} else if size == -1 {
			return nil, nil
		} else {
			return t.readBulkString(size)
		}
	case ARRAY:
		if size, err := t.getCount([]byte(line)); err != nil {
			return nil, err
		} else if size == -1 {
			return nil, nil
		}
		return t.readArray(line)
	default:
		return nil, errors.New("bad element")
	}
}

func (t *Talker) readLine() (string, error) {
	p, err := t.br.ReadSlice('\n')
	if err != nil {
		return "", err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return "", errors.New("bad response line terminator")
	}
	return string(p[:i]), nil
}

func (t *Talker) readBulkString(size int) (string, error) {
	size = size + 2
	buffer, err := t.br.Peek(size)
	t.br.Discard(size)
	return strings.TrimRight(string(buffer), SYM_CRLF), err
}

func (t *Talker) getCount(line []byte) (int, error) {
	return strconv.Atoi(string(line[1:]))
}

func (t *Talker) readArray(line string) ([]interface{}, error) {
	// Get number of array elements.
	count, err := t.getCount([]byte(line))
	if err != nil {
		return nil, err
	}
	// Read `count` number of RESP objects in the array.
	array := make([]interface{}, count, count)
	for i := 0; i < count; i++ {
		buf, err := t.readObject()
		if err != nil {
			return nil, err
		}
		array[i] = buf
	}

	return array, nil
}
