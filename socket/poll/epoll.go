package poll

import (
	"fmt"
	"github.com/panjf2000/gnet/errors"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
)

const (
	EPOLL_EVENT_SIZE = 128
)

type Poller struct {
	fd int
}

func InitPoller() (*Poller, error) {
	fd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return nil, os.NewSyscallError("epoll_create1", err)
	}

	poller := &Poller{
		fd: fd,
	}

	return poller, nil
}

func (poller *Poller) formatEpollEvent(fd int, evt string) (*syscall.EpollEvent, error) {
	e := &syscall.EpollEvent{
		Fd: int32(fd),
	}

	e.Events = -syscall.EPOLLET

	switch evt {
	case "r":
		e.Events |= syscall.EPOLLIN
	case "w":
		e.Events |= syscall.EPOLLOUT
	case "rw":
		e.Events |= syscall.EPOLLIN | syscall.EPOLLOUT
	default:
		return nil, fmt.Errorf("unknow epoll event type")
	}

	return e, nil
}

func (poller *Poller) Add(fd int, evt string) error {
	e, err := poller.formatEpollEvent(fd, evt)
	if err != nil {
		return err
	}

	return os.NewSyscallError("epoll_ctl add", syscall.EpollCtl(poller.fd, syscall.EPOLL_CTL_ADD, fd, e))
}

func (poller *Poller) Mod(fd int, evt string) error {
	e, err := poller.formatEpollEvent(fd, evt)
	if err != nil {
		return err
	}

	return os.NewSyscallError("epoll_ctl mod", syscall.EpollCtl(poller.fd, syscall.EPOLL_CTL_MOD, fd, e))
}

func (poller *Poller) Del(fd int) error {
	return os.NewSyscallError("epoll_ctl del", syscall.EpollCtl(poller.fd, syscall.EPOLL_CTL_DEL, fd, nil))
}

func (poller *Poller) Close() error {
	return os.NewSyscallError("close", syscall.Close(poller.fd))
}

func (poller *Poller) Polling(eventHandler func(fd int32, evts uint32)) error {
	events := make([]syscall.EpollEvent, EPOLL_EVENT_SIZE)
	for {
		n, err := syscall.EpollWait(poller.fd, events, 0)
		if n < 0 && err == syscall.EINTR {
			continue
		}
		if err != nil {
			return os.NewSyscallError("epoll_wait", err)
		}

		for i := 0; i < n; i++ {
			eventHandler(events[i].Fd, events[i].Events)
		}

		for i := 0; i < n; i++ {
			if fd := int(events[i].Fd); fd != poller.fd {
				switch err = callback(fd, events[i].Events); err {
				case nil:
				case errors.ErrAcceptSocket, errors.ErrServerShutdown:
					return err
				default:
					logging.DefaultLogger.Warnf("Error occurs in event-loop: %v", err)
				}
			} else {
				_, _ = unix.Read(poller.fd, p.wfdBuf)
			}
		}
	}
}
