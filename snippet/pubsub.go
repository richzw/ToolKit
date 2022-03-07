package snippet

import (
	"fmt"
	"time"
)

type Broker struct {
	stopCh    chan struct{}
	publishCh chan interface{}
	subCh     chan chan interface{}
	unsubCh   chan chan interface{}
}

func NewBroker() *Broker {
	return &Broker{
		stopCh:    make(chan struct{}),
		publishCh: make(chan interface{}, 1),
		subCh:     make(chan chan interface{}, 1),
		unsubCh:   make(chan chan interface{}, 1),
	}
}

func (b *Broker) Start() {
	subs := map[chan interface{}]struct{}{}
	for {
		select {
		case <-b.stopCh:
			for msgCh := range subs {
				close(msgCh)
			}
			return
		case msgCh := <-b.subCh:
			subs[msgCh] = struct{}{}
		case msgCh := <-b.unsubCh:
			delete(subs, msgCh)
		case msg := <-b.publishCh:
			for msgCh := range subs {
				// msgCh is buffered, use non-blocking send to protect the broker:
				select {
				case msgCh <- msg:
				default:
				}
			}
		}
	}
}

func (b *Broker) Stop() {
	close(b.stopCh)
}

func (b *Broker) Subscribe() chan interface{} {
	msgCh := make(chan interface{}, 5)
	b.subCh <- msgCh
	return msgCh
}

func (b *Broker) Unsubscribe(msgCh chan interface{}) {
	b.unsubCh <- msgCh
}

func (b *Broker) Unsubscribe1(msgCh chan interface{}) {
	b.unsubCh <- msgCh
	close(msgCh)
}

func (b *Broker) Publish(msg interface{}) {
	b.publishCh <- msg
}

func main() {
	b := NewBroker()
	go b.Start()

	clientFunc := func(id int) {
		msgCh := b.Subscribe()
		for msg := range msgCh {
			fmt.Printf("Client %d got message: %v \n", id, msg)
		}
	}
	for i := 0; i < 3; i++ {
		go clientFunc(i)
	}

	go func() {
		for msgId := 0; ; msgId++ {
			b.Publish(fmt.Sprintf("msg# %d", msgId))
			time.Sleep(300 * time.Millisecond)
		}
	}()

	time.Sleep(time.Second)
}
