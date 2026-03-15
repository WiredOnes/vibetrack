package app

import (
	"io/fs"

	api "github.com/WiredOnes/vibetrack/backend/api/http/v1"
	"github.com/WiredOnes/vibetrack/backend/internal/config"
	"github.com/WiredOnes/vibetrack/backend/internal/db"
	"github.com/WiredOnes/vibetrack/backend/internal/environment"
	"github.com/WiredOnes/vibetrack/backend/internal/http"
	"github.com/WiredOnes/vibetrack/backend/internal/logic"
	"github.com/WiredOnes/vibetrack/backend/internal/state"
	"github.com/WiredOnes/vibetrack/backend/internal/telemetry"
	"github.com/benbjohnson/clock"
	"github.com/leshless/golibrary/graceful"
	"github.com/leshless/golibrary/interrupt"
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
	Controller logic.Controller
}

// @PublicValueInstance
type Actions struct{}

// @PublicValueInstance
type Handlers struct {
	// HTTP
	HTTPHandler api.StrictServerInterface
}

// @PublicValueInstance
type Ports struct {
	HTTP http.Port
}
