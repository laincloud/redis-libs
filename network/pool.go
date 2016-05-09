package network

import (
	"errors"
	"github.com/mijia/sweb/log"
	"github.com/laincloud/redis-libs/utils"
	"net"
	"sync"
	"time"
)

var (
	errConnLimited = errors.New("-Error connection limited\r\n")
	errPoolClosed  = errors.New("-Error pool closed\r\n")
)

type dialer func() (net.Conn, error)

type connectOption struct {
	readTimeOutSec time.Duration
	wrteTimeOutSec time.Duration
	bufferSize     int
}

func NewConnectOption(readTimeOut, wrteTimeOut, bs int) *connectOption {
	return &connectOption{
		readTimeOutSec: time.Duration(readTimeOut) * time.Second,
		wrteTimeOutSec: time.Duration(wrteTimeOut) * time.Second,
		bufferSize:     bs,
	}
}

type Pool struct {
	co         *connectOption
	idles      *utils.Queue
	_dialer    dialer
	activeSize int
	closed     bool

	idleTimeOutSec time.Duration
	connStateTest  connStateTestFunc

	maxActive int
	maxIdle   int
	mu        *sync.RWMutex
	connMu    *sync.Mutex
}

type idleConn struct {
	cn       *Conn
	idleTime time.Time
}

type connStateTestFunc func(c *Conn) bool

func NewConnectionPool(co *connectOption, _dialer dialer, maxActive, maxIdle, idleTimeOutSec int) *Pool {
	idles := utils.NewQueue()
	return &Pool{
		co:             co,
		idles:          idles,
		_dialer:        _dialer,
		activeSize:     0,
		closed:         false,
		idleTimeOutSec: time.Duration(idleTimeOutSec) * time.Second,
		maxIdle:        maxIdle,
		maxActive:      maxActive,
		mu:             &sync.RWMutex{},
		connMu:         &sync.Mutex{},
	}
}

func (p *Pool) SetConnStateTest(cstf connStateTestFunc) {
	p.connStateTest = cstf
}

func (p *Pool) SetDialer(newDialer dialer) {
	p._dialer = newDialer
}

func (p *Pool) newConnect() (*Conn, error) {
	if p.active() {
		if conn, err := p._dialer(); err != nil {
			p.release()
			return nil, err
		} else {
			return NewConnect(conn, p.co), nil
		}
	}
	return nil, errConnLimited
}

func (p *Pool) get() *Conn {
	p.connMu.Lock()
	defer p.connMu.Unlock()
	for {
		c := p.idles.Pop()
		if c == nil {
			return nil
		}
		conn, _ := c.(*idleConn)
		if conn.idleTime.Add(p.idleTimeOutSec).After(time.Now()) {
			return conn.cn
		}
		p.closeConn(conn.cn)
	}

}

func (p *Pool) put(c *Conn) {
	p.connMu.Lock()
	defer p.connMu.Unlock()
	if p.idles.Length() < p.maxIdle {
		p.idles.Push(&idleConn{cn: c, idleTime: time.Now()})
	} else {
		p.closeConn(c)
	}
}

func (p *Pool) limited() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.activeSize <= p.maxActive {
		return false
	}
	return true
}

func (p *Pool) active() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.activeSize < p.maxActive {
		p.activeSize++
		return true
	}
	return false
}

func (p *Pool) release() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.activeSize > 0 {
		p.activeSize--
		return true
	}
	return false
}

func (p *Pool) fetchConn() (*Conn, error) {
	for {
		c := p.get()
		if c == nil {
			break
		}
		// test state of c
		if p.connStateTest != nil {
			b := p.connStateTest(c)
			if !b {
				p.closeConn(c)
				continue
			}
		}
		return c, nil
	}
	return p.newConnect()
}

func (p *Pool) closeConn(c *Conn) {
	c.Close()
	p.release()
}

func (p *Pool) FetchConn() (*Conn, error) {
	p.mu.Lock()
	if p.closed {
		return nil, errPoolClosed
	}
	p.mu.Unlock()
	c, err := p.fetchConn()
	log.Debugf("%d::%d::%d\n", p.idles.Length(), p.activeSize, p.maxActive)
	return c, err
}

func (p *Pool) Close() {
	p.mu.Lock()
	p.closed = true
	p.activeSize -= p.idles.Length()
	p.mu.Unlock()
	for {
		c := p.get()
		if c == nil {
			return
		}
		c.Close()
	}
}

func (p *Pool) Finished(c *Conn) {
	if c == nil {
		return
	}
	if c.err != nil || p.closed {
		p.closeConn(c)
	} else {
		p.put(c)
	}
}
