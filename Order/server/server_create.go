package server

import (
	pb "Order/grpc"
	"Order/kafka"
	"Order/storage"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

var (
	port = flag.Int("port", 8000, "The Server port")
)

type Server struct {
	pb.UnimplementedOrderServer
	storage         *storage.StorageOrder
	kafkaConnection *kafka.KafkaOrderConnection
	mutex           sync.Mutex
}

func (server *Server) New(storageOrder *storage.StorageOrder, kafkaConnection *kafka.KafkaOrderConnection) {

	server.storage = storageOrder
	server.kafkaConnection = kafkaConnection

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServer(grpcServer, server)
	log.Printf("Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
