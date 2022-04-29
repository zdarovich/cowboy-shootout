package main

import (
	"context"
	pbCowboy "github.com/zdarovich/cowboy_shooters/api/cowboy"
	"github.com/zdarovich/cowboy_shooters/internal/cowboy"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"sync"
)

var (
	logServiceAddr = flag.String("logServiceAddr", os.Getenv("LOG_ADDR"), "The log service addr")
	name           = flag.String("name", os.Getenv("NAME"), "Cowboy name")
	health         = flag.String("health", os.Getenv("HEALTH"), "Cowboy health")
	damage         = flag.String("damage", os.Getenv("DAMAGE"), "Cowboy damage")
	port           = flag.String("port", os.Getenv("PORT"), "Cowboy port")
)

// Cowboy runner and server listener starts here
func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cowboyGuy, err := cowboy.NewCowboy(*name, *health, *damage, *logServiceAddr)
	if err != nil {
		log.Fatalf("failed to create cowbow listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	server := CowboyServer{cowboy: cowboyGuy}
	pbCowboy.RegisterCowboyServer(grpcSrv, &server)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s := <-cowboyGuy.IsDead()
		log.Printf("got signal %v, attempting graceful shutdown", s)
		grpcSrv.GracefulStop()
		wg.Done()
	}()

	go cowboyGuy.Run()

	log.Printf("starting grpc server under port %d", port)
	err = grpcSrv.Serve(lis)
	if err != nil {
		log.Fatalf("could not serve: %v", err)
	}
	wg.Wait()
	log.Println("clean shutdown")
}

type CowboyServer struct {
	pbCowboy.UnsafeCowboyServer
	cowboy *cowboy.Service
}

// HandleShoot decrements cowboy health by damage number
func (c *CowboyServer) HandleShoot(ctx context.Context, request *pbCowboy.ShootRequest) (*pbCowboy.ShootResponse, error) {
	health := c.cowboy.GetHealth()
	if health == 0 {
		return &pbCowboy.ShootResponse{Msg: fmt.Sprintf("%s is dead already", c.cowboy.Name)}, nil
	}
	c.cowboy.GetShoot(int(request.Damage))
	return &pbCowboy.ShootResponse{Msg: fmt.Sprintf("%s was shot with %d damage", c.cowboy.Name, c.cowboy.Damage)}, nil
}
