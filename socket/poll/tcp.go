package poll

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
)

const SYN_BACKLOG_SIZE = 128

func NewTCPSocket(network, addr string, canReusePort bool) (int, syscall.Sockaddr, error) {
	var sa syscall.Sockaddr
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return 0, nil, fmt.Errorf("resolve addr err: %v", err)
	}

	netFamily := syscall.AF_INET
	if tcpAddr.IP.To4() == nil {
		netFamily = syscall.AF_INET6
	}

	syscall.ForkLock.Lock()
	// close the file if executing a new processes (O_CLOEXEC)
	// this is important to not copy file descriptors between processes unexpectedly
	fd, err := syscall.Socket(netFamily, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, syscall.IPPROTO_TCP)
	if err == nil {
		syscall.CloseOnExec(fd)
	}
	syscall.ForkLock.Unlock()

	defer func() {
		if err != nil {
			_ = syscall.Close(fd)
		}
	}()

	if err = syscall.SetNonblock(fd, true); err != nil {
		return -1, nil, os.NewSyscallError("setnonblock", err)
	}

	switch network {
	case "tcp":
		sockaddr := &syscall.SockaddrInet4{}
		sockaddr.Port = tcpAddr.Port
		sa = sockaddr
	case "tcp4":
		sockaddr := &syscall.SockaddrInet4{}
		sockaddr.Port = tcpAddr.Port
		copy(sockaddr.Addr[:], tcpAddr.IP.To4())
		sa = sockaddr
	case "tcp6":
		sockaddr := &syscall.SockaddrInet6{}
		copy(sockaddr.Addr[:], tcpAddr.IP.To16())
		if tcpAddr.Zone != "" {
			var iface *net.Interface
			iface, err = net.InterfaceByName(tcpAddr.Zone)
			if err != nil {
				return 0, nil, fmt.Errorf("parse TCPAddr.Zone err: %v", err)
			}
			sockaddr.ZoneId = uint32(iface.Index)
		}
		sockaddr.Port = tcpAddr.Port
		netFamily = syscall.AF_INET6
	default:
		return 0, nil, fmt.Errorf("not support network")
	}

	if canReusePort {
		if err = os.NewSyscallError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, SO_REUSEPORT, 1)); err != nil {
			return 0, nil, err
		}
	}

	if err = os.NewSyscallError("bind", syscall.Bind(fd, sa)); err != nil {
		return 0, nil, err
	}

	if err = os.NewSyscallError("listen", syscall.Listen(fd, SYN_BACKLOG_SIZE)); err != nil {
		return 0, nil, err
	}

	return fd, sa, nil
}

func IsTCP(network string) bool {
	switch strings.ToLower(network) {
	case "tcp", "tcp4", "tcp6":
		return true
	}

	return false
}