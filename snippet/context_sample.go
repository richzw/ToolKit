package snippet

// blocking with context
type Database struct {
	cancel func()
	wait   func() error
}

func (db *Database) Setup() error {
	ctx, cancel := context.WithCancel(context.Background())
	g, gCtx := errgroup.WithContext(ctx)

	db.cancel = cancel
	db.wait = g.Wait

	for {
		select {
		case <-ctx.Done():
			return ctx.Err() // Depending on our business logic,
			//   we may or may not want to return a ctx error:
			//   https://pkg.go.dev/context#pkg-variables
		}
	}
}

func (db *Database) Shutdown() error {
	db.cancel()
	return db.wait()
}
