package telemetry

import (
	"context"
	"fmt"
	"slices"

	"github.com/leshless/golibrary/xmaps"
	"github.com/leshless/golibrary/xslices"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	ErrorLabel       = "error"
	ServiceLabel     = "service"
	HostLabel        = "host"
	EnvironmentLabel = "environment"
	CallLabel        = "call"
	MethodLabel      = "method"
	EndpointLabel    = "endpoint"
	StatusCodeLabel  = "status_code"
	StatusLabel      = "status"
	JobNameLabel     = "job_name"
	SuccessfulLabel  = "successful"
)

// @PrivateValueInstance
type label struct {
	key   string
	value any
}

func Error(err error) label {
	return newLabel(ErrorLabel, err.Error())
}

func Service(value string) label {
	return newLabel(ServiceLabel, value)
}

func Host(value string) label {
	return newLabel(HostLabel, value)
}

func Environment(value string) label {
	return newLabel(EnvironmentLabel, value)
}

func Call(value string) label {
	return newLabel(CallLabel, value)
}

func Method(value string) label {
	return newLabel(MethodLabel, value)
}

func Endpoint(value string) label {
	return newLabel(EndpointLabel, value)
}

func StatusCode(value int) label {
	return newLabel(StatusCodeLabel, value)
}

func Status(value string) label {
	return newLabel(StatusLabel, value)
}

func JobName(value string) label {
	return newLabel(JobNameLabel, value)
}

func Successful(value bool) label {
	return newLabel(SuccessfulLabel, value)
}

func Any(key string, value any) label {
	return newLabel(key, value)
}

func allLabels(ctx context.Context, labels []label) []label {
	ctxLabels := labelsFromContext(ctx)

	for _, label := range labels {
		ctxLabels[label.key] = label.value
	}

	return xmaps.UnzipFunc(ctxLabels, func(key string, value any) label {
		return newLabel(key, value)
	})
}

func labelKeys(labels []label) []string {
	labelKeys := xslices.Map(labels, func(l label) string {
		return l.key
	})
	slices.Sort(labelKeys)

	return labelKeys
}

func labelsToZap(labels []label) []zap.Field {
	return xslices.Map(labels, func(label label) zap.Field {
		return zap.Any(label.key, label.value)
	})
}

func labelsToProm(labels []label) prometheus.Labels {
	return xmaps.ZipFunc(labels, func(l label) (string, string) {
		return l.key, fmt.Sprintf("%v", l.value)
	})
}
