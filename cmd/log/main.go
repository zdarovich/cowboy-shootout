package main

import (
	"context"
	logService "github.com/zdarovich/cowboy_shooters/internal/log"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/zdarovich/cowboy_shooters/api/log"
	"google.golang.org/grpc"
)

var (
	port = flag.String("port", os.Getenv("PORT"), "The server port")
)

// Grpc server for logs starts here
func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	logSvc, err := logService.NewLog(".")
	if err != nil {
		log.Fatalf("failed to init log: %v", err)
	}
	defer logSvc.Close()
	s := grpc.NewServer()
	pb.RegisterLogServer(s, &server{Log: logSvc})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type server struct {
	pb.UnsafeLogServer
	Log *logService.Log
}

// HandleProduce saves logs to storage
func (s *server) HandleProduce(ctx context.Context, req *pb.ProduceRequest) (*pb.ProduceResponse, error) {
	log.Println(req.Record)
	off, err := s.Log.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &pb.ProduceResponse{Offset: off}, err
}
