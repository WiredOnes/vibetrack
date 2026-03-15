package app

import (
	"io/fs"

	"github.com/benbjohnson/clock"
	"github.com/leshless/golibrary/graceful"
	"github.com/leshless/golibrary/interrupt"
	cubpb "github.com/leshless/pet/cub/api/grpc/v1"
	"github.com/leshless/pet/cub/internal/config"
	"github.com/leshless/pet/cub/internal/db"
	"github.com/leshless/pet/cub/internal/environment"
	"github.com/leshless/pet/cub/internal/grpc"
	"github.com/leshless/pet/cub/internal/http"
	"github.com/leshless/pet/cub/internal/job"
	"github.com/leshless/pet/cub/internal/logic/health"
	"github.com/leshless/pet/cub/internal/state"
	"github.com/leshless/pet/cub/internal/telemetry"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// @PublicValueInstance
type Primitives struct {
	Clock       clock.Clock
	Interrupter interrupt.Interrupter
	FS          fs.FS
}

// @PublicValueInstance
type Dependencies struct {
	GracefulRegistrator graceful.Registrator
	GracefulTerminator  graceful.Terminator
	ConfigHolder        config.Holder
	EnvironmentHolder   environment.Holder
	Logger              telemetry.Logger
	Registry            telemetry.Registry
	Telemetry           telemetry.Telemetry
}

// @PublicValueInstance
type Clients struct {
	DB db.Client
}

// @PublicValueInstance
type Adapters struct {
	// DB
	PingDB db.PingAdapter
	TXDB   db.TxAdapter
	// State
	HealthState state.HealthAdapter
}

// @PublicValueInstance
type Controllers struct {
	Health health.Controller
}

// @PublicValueInstance
type Actions struct{}

// @PublicValueInstance
type Handlers struct {
	// GRPC
	HealthGRPC healthpb.HealthServer
	PingGRPC   cubpb.PingServer
	// HTTP
	HealthHTTP *http.HealthHandler
	// Jobs
	UpdateHealthStatusJob job.UpdateHealthStatus
}

// @PublicValueInstance
type Ports struct {
	GRPC grpc.Port
	HTTP http.Port
	Job  job.Port
}
