package poll

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
)

const SO_REUSEPORT = 0xf

func NewUDPSocket(network, addr string, canReusePort bool) (int, syscall.Sockaddr, error) {
	var sa syscall.Sockaddr
	udpAddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		return 0, nil, fmt.Errorf("resolve addr err: %v", err)
	}

	netFamily := syscall.AF_INET
	if udpAddr.IP.To4() == nil {
		netFamily = syscall.AF_INET6
	}

	syscall.ForkLock.Lock()
	// close the file if executing a new processes (O_CLOEXEC)…
	// this is important to not copy file descriptors between processes unexpectedly
	fd, err := syscall.Socket(netFamily, syscall.SOCK_DGRAM|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, syscall.IPPROTO_UDP)
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
	case "udp":
		sockaddr := &syscall.SockaddrInet4{}
		sockaddr.Port = udpAddr.Port
		sa = sockaddr
	case "udp4":
		sockaddr := &syscall.SockaddrInet4{}
		sockaddr.Port = udpAddr.Port
		copy(sockaddr.Addr[:], udpAddr.IP.To4())
		sa = sockaddr
	case "udp6":
		sockaddr := &syscall.SockaddrInet6{}
		copy(sockaddr.Addr[:], udpAddr.IP.To16())
		if udpAddr.Zone != "" {
			var iface *net.Interface
			iface, err = net.InterfaceByName(udpAddr.Zone)
			if err != nil {
				return 0, nil, fmt.Errorf("parse UDPAddr.Zone err: %v", err)
			}
			sockaddr.ZoneId = uint32(iface.Index)
		}
		sockaddr.Port = udpAddr.Port
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

	return fd, sa, nil
}

func IsUDP(network string) bool {
	switch strings.ToLower(network) {
	case "udp", "udp4", "udp6":
		return true
	}

	return false
}
