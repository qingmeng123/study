package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/features/proto/echo"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

var (
	addr        = flag.String("addr", "localhost:50052", "the address to connect to")
	retryPolicy = `{
		"methodConfig": [{
		  "name": [{"service": "grpc.examples.echo.Echo"}],
		  "waitForReady": true,
		  "retryPolicy": {
			  "MaxAttempts": 5,
			  "InitialBackoff": "1s",
			  "MaxBackoff": "1s",
			  "BackoffMultiplier": 1.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
		  }
		}]}`
)

func retryDial() (*grpc.ClientConn, error) {
	return grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultServiceConfig(retryPolicy), grpc.WithUnaryInterceptor(LoggingInterceptor))
}

func main() {
	flag.Parse()
	conn, err := retryDial()
	if err != nil {
		log.Fatalf("did not connect:%v", err)
	}
	defer func() {
		if e := conn.Close(); e != nil {
			log.Printf("failed to close connection:%s", e)
		}
	}()
	c := pb.NewEchoClient(conn)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		reply, err := c.UnaryEcho(ctx, &pb.EchoRequest{Message: "Try and Success"})
		if err != nil {
			log.Fatalf("UnaryEcho error:%v", err)
		}
		log.Printf("UnaryEcho reply:%v", reply)
	}

}

func LoggingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	startTime := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	endTime := time.Now()

	if err != nil {
		statusErr, _ := status.FromError(err)
		log.Printf("gRPC method: %s, status: %s, duration: %v", method, statusErr.Message(), endTime.Sub(startTime))
	} else {
		log.Printf("gRPC method: %s, status: OK, duration: %v", method, endTime.Sub(startTime))
	}

	return err
}
