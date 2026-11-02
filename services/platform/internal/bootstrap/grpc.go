package bootstrap

import (
	"github.com/hycat-team/littlecafe-backend/contracts/grpcx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// GRPCServerRegistrar is implemented by business modules that expose gRPC
// services. Modules do not import bootstrap: the interface is structural.
type GRPCServerRegistrar interface {
	RegisterGRPC(server *grpc.Server)
}

// NewGRPCServer builds the gRPC server with health checking and reflection
// enabled, then lets each registrar mount its contract services.
func NewGRPCServer(registrars []GRPCServerRegistrar) (*grpc.Server, error) {
	options, err := grpcx.ServerOptionsFromEnv()
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer(options...)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)

	for _, registrar := range registrars {
		registrar.RegisterGRPC(server)
	}

	reflection.Register(server)
	return server, nil
}
