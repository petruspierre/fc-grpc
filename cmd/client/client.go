package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/petruspierre/go-grpc/pb"
	"google.golang.org/grpc"
)

func main() {
	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer connection.Close()

	client := pb.NewUserServiceClient(connection)

	// AddUser(client)
	// AddUserVerbose(client)
	// AddUsers(client)
	AddUserStreamBoth(client)
}

func AddUser(client pb.UserServiceClient) {
	req := &pb.User{
		Id:    "0",
		Name:  "Petrus",
		Email: "ppierre@trans-stat.com",
	}

	res, err := client.AddUser(context.Background(), req)

	if err != nil {
		log.Fatalf("could not make gRPC request: %v", err)
	}

	fmt.Println(res)
}

func AddUserVerbose(client pb.UserServiceClient) {
	req := &pb.User{
		Id:    "0",
		Name:  "Petrus",
		Email: "ppierre@trans-stat.com",
	}

	responseStream, err := client.AddUserVerbose(context.Background(), req)

	if err != nil {
		log.Fatalf("could not make gRPC request: %v", err)
	}

	for {
		stream, err := responseStream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("could not receive the stream message: %v", err)
		}

		fmt.Println("Status: ", stream.Status, " - ", stream.GetUser())
	}
}

func AddUsers(client pb.UserServiceClient) {
	reqs := []*pb.User{
		&pb.User{
			Id:    "p1",
			Name:  "Petrus 1",
			Email: "p1@p.com",
		},
		&pb.User{
			Id:    "p2",
			Name:  "Petrus 2",
			Email: "p2@p.com",
		},
		&pb.User{
			Id:    "p3",
			Name:  "Petrus 3",
			Email: "p3@p.com",
		},
	}

	stream, err := client.AddUsers(context.Background())

	if err != nil {
		log.Fatalf("could not make gRPC request: %v", err)
	}

	for _, req := range reqs {
		stream.Send(req)
		time.Sleep(time.Second * 3)
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("could not receive the response: %v", err)
	}

	fmt.Println(res)
}

func AddUserStreamBoth(client pb.UserServiceClient) {
	stream, err := client.AddUserStreamBoth(context.Background())

	if err != nil {
		log.Fatalf("could not make gRPC request: %v", err)
	}

	reqs := []*pb.User{
		&pb.User{
			Id:    "p1",
			Name:  "Petrus 1",
			Email: "p1@p.com",
		},
		&pb.User{
			Id:    "p2",
			Name:  "Petrus 2",
			Email: "p2@p.com",
		},
		&pb.User{
			Id:    "p3",
			Name:  "Petrus 3",
			Email: "p3@p.com",
		},
	}

	wait := make(chan int)

	go func() {
		for _, req := range reqs {
			fmt.Println("Sending user: ", req.Name)
			stream.Send(req)
			time.Sleep(time.Second * 2)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("could not receive the response: %v", err)
				break
			}

			fmt.Printf("Receiving user %v with status %v \n", res.GetUser().GetName(), res.GetStatus())
		}
		close(wait)
	}()

	<-wait
}
