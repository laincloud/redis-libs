package network

import (
	"github.com/mijia/sweb/log"
	"syscall"
	"time"
)

func SyscallWrite(fd int, b []byte, aeBufferSize int) error {
	from := 0
	to := 0
	respSize := len(b)
	for {
		if from+aeBufferSize < respSize {
			to = from + aeBufferSize
		} else {
			to = respSize
		}
		if n, err := syscall.Write(fd, b[from:to]); err != nil {
			log.Error(err.Error())
			if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return err
		} else {
			from += n
		}
		if from == respSize {
			break
		}
	}
	return nil
}

func SyscallRead(fd int, aeBufferSize int) ([]byte, error) {
	buf := make([]byte, aeBufferSize)
	msg := make([]byte, 0)
	for {
		nbytes, err := syscall.Read(fd, buf)
		msg = append(msg, buf[:nbytes]...)
		if err != nil {
			return nil, err
		} else if nbytes < aeBufferSize {
			break
		}

	}
	return msg, nil
}
