package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"study/hello_client/pb"
	"testing"
	"time"
)

func Test_MustInitClient(t *testing.T) {
	config := Config{
		Domain:      "www.duryun.xyz",
		Endpoints:   []string{"127.0.0.1:8972", "127.0.0.1:8973"},
		MaxAttempts: 0,
		Opts:        make([]grpc.DialOption, 0),
	}
	config.Opts = append(config.Opts, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn := MustInitClient(config)

	c := pb.NewGreeterClient(conn)

	// 执行10次RPC调用查看是否轮询
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		time.Sleep(time.Second)
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "hello"})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s", r.GetReply())
	}
}
