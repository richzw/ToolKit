package snippet

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func grpcShutdown() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gSrv := greeterServer{}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpc := grpc.NewServer()
	pb.RegisterGreeterServer(grpc, &gSrv)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s := <-sigCh
		log.Printf("got signal %v, attempting graceful shutdown", s)
		cancel()
		grpc.GracefulStop()
		// grpc.Stop() // leads to error while receiving stream response: rpc error: code = Unavailable desc = transport is closing
		wg.Done()
	}()

	log.Println("starting grpc server")
	err = grpc.Serve(lis)
	if err != nil {
		log.Fatalf("could not serve: %v", err)
	}

	wg.Wait()
	log.Println("clean shutdown")
}

func httpShutDown() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr: ":8000",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			for {
				select {
				case <-ctx.Done():
					fmt.Println("Graceful handler exit")
					w.WriteHeader(http.StatusOK)
					return
				case <-time.After(1 * time.Second):
					fmt.Println("Hello in a loop")
				}
			}
		}),
		BaseContext: func(_ net.Listener) context.Context {
			return mainCtx
		},
	}
	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}
}
