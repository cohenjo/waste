package grpc


//go:generate protoc -I ./waste --go_out=plugins=grpc:./waste ./waste/waste.proto

import (
	"context"
	"github.com/rs/zerolog/log"
	"net"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "github.com/cohenjo/waste/go/grpc/waste"
	"github.com/cohenjo/waste/go/config"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"

)

var (
	messages map[string]string
)


// server is used to implement waste....Server.
type server struct{
	
}

func (s *server) Status(filter *pb.Filter, toClient pb.Waste_StatusServer) error {
	cngStatus := &pb.ChangeStatus{
		ChangeState:    pb.State_RUNNING,
		Message: "test",
		Uuid: "XXX-XXX-XXX-XXX",
	}
	err := toClient.Send(cngStatus)
	return err
}

// SendMessage implements helloworld.GreeterServer
func (s *server) RunChange(ctx context.Context,in *pb.Change) (*pb.ChangeStatus, error) {
	return &pb.ChangeStatus{
		ChangeState:    pb.State_RUNNING,
		Message: "test",
		Uuid: "XXX-XXX-XXX-XXX",
	}, nil
}


func Serve() {

	// chceck if enable
	if !config.Config.GrpcEnable {
		log.Warn().Msg("GRPC is DISABLE , will not start")
		return
	}

	log.Debug().Msg("Starting GRPC server")
	// open port for grpc endpoint
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Config.GrpcListeningPort))
	if err != nil {
		log.Fatal().Err(err).Msgf("GRPC Unable to listen on port: %d", config.Config.GrpcListeningPort)
	}

	// create grpc server with OpenTracing middelware
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(
		otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
		grpc.StreamInterceptor(
			otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer())))
	
	pb.RegisterWasteServer(grpcServer, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	// serve it
	go func() {
		defer log.Info().Msg("GRPC server stopped")

		log.Info().Msgf("GRPC server started on port: %d", config.Config.GrpcListeningPort)
		serverErr := grpcServer.Serve(lis)
		if serverErr != nil {
			log.Fatal().Err(serverErr).Msg("GRPC server failed")
		}
	}()
}
