package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/leshless/pet/cub/internal/config"
	"github.com/leshless/pet/cub/internal/telemetry"
	"google.golang.org/grpc"
)

type Port interface {
	Run(ctx context.Context) error
}

// @PublicPointerInstance
type port struct {
	telemetry.Telemetry
	configHolder config.Holder
	grpcServer   *grpc.Server
}

var _ Port = (*port)(nil)

func (p *port) Run(ctx context.Context) error {
	p.Logger.Info(ctx, "starting grpc server")

	config := p.configHolder.Config().GRPC

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		p.Logger.Error(ctx, "failed to start grpc server", telemetry.Error(err))

		return fmt.Errorf("starting grpc server: %w", err)
	}

	err = p.grpcServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("grpc server stopped with error: %w", err)
	}

	p.Logger.Warn(ctx, "grpc server stopped")

	return nil
}
