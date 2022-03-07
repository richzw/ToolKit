package snippet

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	log.Println("starting grpc server")
	err = grpc.Serve(lis)
	if err != nil {
		log.Fatalf("could not serve: %v", err)
	}

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

	wg.Wait()
	log.Println("clean shutdown")
}
