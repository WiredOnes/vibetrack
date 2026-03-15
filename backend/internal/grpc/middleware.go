package grpc

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/leshless/pet/cub/internal/model"
	"github.com/leshless/pet/cub/internal/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func recoveryMiddleware(tel telemetry.Telemetry) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				tel.Logger.Error(ctx, "caught panic in unary grpc call", telemetry.Any("panic_value", r))
				err = errorFromModel(model.NewInternalError())
			}
		}()

		return handler(ctx, req)
	}
}

func errorMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		response, err := handler(ctx, req)
		if err != nil {
			return nil, errorFromModel(err)
		}

		return response, nil
	}
}

func telemetryMiddleware(clock clock.Clock, tel telemetry.Telemetry) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		ts := clock.Now()

		ctx = telemetry.ContextWith(ctx, telemetry.Call(info.FullMethod))

		tel.Logger.Info(ctx, "recieved unary grpc call")

		response, err := handler(ctx, req)

		latency := time.Since(ts)

		ctx = telemetry.ContextWith(
			ctx,
			telemetry.Status(status.Code(err).String()),
			telemetry.Any("latency", latency),
		)

		tel.Registry.Counter(ctx, telemetry.GRPCCallsTotal).Inc()
		tel.Registry.Summary(ctx, telemetry.GRPCCallDurationSeconds).Observe(latency.Seconds())

		if err != nil {
			tel.Logger.Warn(
				ctx,
				"processed unary grpc call with error",
				telemetry.Error(err),
			)
		} else {
			tel.Logger.Info(ctx, "sucessfully processed unary grpc call")
		}

		return response, err
	}
}
