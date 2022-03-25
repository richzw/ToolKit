package snippet

import (
	"io"
	"net"
	"sync"
)

// 1. Common sense to add mutex to conns
type Mux struct {
	mu    sync.Mutex
	conns map[net.Addr]net.Conn
}

func (m *Mux) Add(conn net.Conn) {
	m.mu.Lock()
	m.conns[conn.RemoteAddr()] = conn
	m.mu.Unlock()
}

func (m *Mux) Remove(addr net.Addr) {
	m.mu.Lock()
	delete(m.conns, addr)
	m.mu.Unlock()
}

func (m *Mux) SendMessage(msg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.conns {
		if _, err := io.WriteString(conn, msg); err != nil {
			return err
		}
	}
	return nil
}

// 2. replace mutex with channels
type MuxWithChan struct {
	add     chan net.Conn
	remove  chan net.Addr
	sendMsg chan string
}

func (m *MuxWithChan) Add(conn net.Conn) {
	m.add <- conn
}

func (m *MuxWithChan) Remove(addr net.Addr) {
	m.remove <- addr
}

func (m *MuxWithChan) SendMessage(msg string) error {
	m.sendMsg <- msg
	return nil
}

func (m *MuxWithChan) loop() {
	conns := map[net.Addr]net.Conn{}

	for {
		select {
		case conn := <-m.add:
			conns[conn.RemoteAddr()] = conn
		case addr := <-m.remove:
			delete(conns, addr)
		case msg := <-m.sendMsg:
			for _, conn := range conns {
				io.WriteString(conn, msg)
			}
		}
	}
}

// 3. functional refactor
type MuxWithFunctional struct {
	ops chan func(map[net.Addr]net.Conn)
}

func (m *MuxWithFunctional) Add(conn net.Conn) {
	m.ops <- func(m map[net.Addr]net.Conn) {
		m[conn.RemoteAddr()] = conn
	}
}

func (m *MuxWithFunctional) Remove(addr net.Addr) {
	m.ops <- func(m map[net.Addr]net.Conn) {
		delete(m, addr)
	}
}

func (m *MuxWithFunctional) SendMessage(msg string) error {
	result := make(chan error, 1)
	m.ops <- func(m map[net.Addr]net.Conn) {
		for _, conn := range m {
			if _, err := io.WriteString(conn, msg); err != nil {
				result <- err
				return
			}
		}
		result <- nil
	}

	return <-result
}

func (m *MuxWithFunctional) loop() {
	conns := map[net.Addr]net.Conn{}

	for op := range m.ops {
		op(conns)
	}
}
