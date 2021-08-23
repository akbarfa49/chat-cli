package main

import (
	"fmt"
	"log"
	pb "msgrpc/messageServ"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)


type server struct{
	pb.UnimplementedMessageServiceServer
	room map[string]map[string]pb.MessageService_OpenMessageChannelServer
	message chan map[string]map[string]string
}

func NewChatServer() *server{
	return &server{
		room: make(map[string]map[string]pb.MessageService_OpenMessageChannelServer,1000),
		message: make(chan map[string]map[string]string, 1000)}
}

func (s *server) Retrieve( stream pb.MessageService_OpenMessageChannelServer) {		
	var lastRoom string
	p, ok := peer.FromContext(stream.Context())
	if !ok{
		return
	}
	for {
		recv, err :=stream.Recv()
		if err != nil{
			delete(s.room[lastRoom], p.Addr.String())
			return
		}
		switch{
		case recv.Room != lastRoom:
			if _, ok :=s.room[recv.Room]; !ok{
				s.room[recv.Room] = make(map[string]pb.MessageService_OpenMessageChannelServer)
			}
			s.room[recv.Room][p.Addr.String()] = stream 
			if _, ok := s.room[lastRoom]; ok {
			delete(s.room[lastRoom], p.Addr.String())
			}
			lastRoom = recv.Room
			break;
		case recv != nil:
			s.message <- map[string]map[string]string{
				recv.Room: {"name":recv.Name,"message":recv.Message, "ip": p.Addr.String()},
			}
			break
		}
	}
}

func (s *server) OpenMessageChannel( stream pb.MessageService_OpenMessageChannelServer) error{
	//ctx := new(context.Context)
	 s.Retrieve(stream)
	
	// for{
	// 	m := <-s.message
	// 	for _, v := range m{
	// 	if err := stream.Send(&pb.MessageResponse{Message: v}); err != nil{
	// 		return err
	// 	}
	// }
	// }
	return nil
}

func (s *server) Send(){
	for{
	for k,v := range <-s.message{
		go func(k string, m map[string]string){
		for j,v:= range s.room[k]{
			if m["ip"] == j{
				continue
			}
			v.Send(&pb.MessageResponse{Name: m["name"],Message: m["message"]})
		}
		}(k,v)
		
	}
}

}

func main() {
	lis, err := net.Listen("tcp", "localhost:9091")
	if err != nil{
		log.Fatalf("Failed to listen: %v", err)
	}
	
	grpcServer := grpc.NewServer()
	s := NewChatServer()
	pb.RegisterMessageServiceServer(grpcServer, s)
	log.Println("listening to localhost")
	go s.Send()
	if err := grpcServer.Serve(lis); err != nil{
		fmt.Println(err)
		grpcServer.Stop()
	}
}