package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "msgrpc/messageServ"

	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx,"localhost:9091", grpc.WithInsecure())
	if err != nil{
		log.Fatalf("Failed connecting gRPC: %v", err)
	}
	defer conn.Close()
	client := pb.NewMessageServiceClient(conn)

	runChat(client)
}

func runChat(client pb.MessageServiceClient){
	stream, err :=client.OpenMessageChannel(context.Background())
	if err != nil {
		log.Fatalf("%v.OpenMessageChannel = _, %v", client, err)
	}
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				fmt.Println(err)
			return
			}
			fmt.Printf("%v : %v\n", in.Name, in.Message)
		}
	}()
	var name string
	var room string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Ready! \n\n")
	for {
		
		if ok :=scanner.Scan(); !ok{
			fmt.Println("exiting")
			os.Exit(0)
			
		}
		msg := scanner.Text()
		fmt.Println()
		switch strings.Split(msg, " ")[0]{
		case "/set":
			name = strings.Join(strings.Split(msg, " ")[1:], " ")
			break
		case "/join":
			room = strings.Join(strings.Split(msg, " ")[1:], " ")
			stream.Send(&pb.MessageRequest{Room: room})
		default:
			if room == ""{
				break
			}
			stream.Send(&pb.MessageRequest{Name: name,Message: msg,Room: room})
			break
		}
	}
}