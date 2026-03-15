package job

import (
	"context"
)

type Job interface {
	Exec(ctx context.Context) error
}
