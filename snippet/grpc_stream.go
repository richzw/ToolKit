package snippet

// One way to close grpc stream on server side
func (m *MyThing) MyBidiServer(stream somepb.Thing_ThingServer) {
	can := make(chan struct{})
	thingRunner := &ThingRunner{
		cancel: can,
		send:   make(chan *somepb.OutgoingMessage, 100),
	}
	m.addRunner(thingRunner) // assume this is mutexed
	defer m.removeRunner(thingRunner)
	go thingRunner.recvLoop(stream)
	go thingRunner.sendLoop(stream)
	// could be select here if you also have a context
	<-can
}

func (t *ThingRunner) recvLoop(stream somepb.Thing_ThingServer) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			break
		}
	}
	t.cancelSafely() // in case client disconnected
}

func (t *ThingRunner) sendLoop(stream somepb.Thing_ThingServer) {
	for msg := range t.send {
		if err := stream.Send(msg); err != nil {
			break
		}
	}
	t.cancelSafely() // in case of network error
}

func (t *ThingRunner) SendMessage(ctx context.Context, msg *somepb.Message) error {
	if t.isCanceled() {
		return ThingClosedError
	}
	select {
	case t.send <- msg:
		// all is well
	case <-ctx.Done():
		return ctx.Err()
	default:
		// the client is too slow -- disconnect it
		t.cancelSafely()
		return ClientTooSlowError
	}
	return nil
}
