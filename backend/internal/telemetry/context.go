package telemetry

import (
	"context"
	"maps"
)

type contextKey struct{}

func ContextWith(ctx context.Context, labels ...label) context.Context {
	ctxLabels := labelsFromContext(ctx)

	for _, label := range labels {
		ctxLabels[label.key] = label.value
	}

	return context.WithValue(ctx, contextKey{}, ctxLabels)
}

func ContextWithout(ctx context.Context, labelKeys ...string) context.Context {
	ctxLabels := labelsFromContext(ctx)

	for _, labelKey := range labelKeys {
		delete(ctxLabels, labelKey)
	}

	return context.WithValue(ctx, contextKey{}, ctxLabels)
}

func labelsFromContext(ctx context.Context) map[string]any {
	ctxValue := ctx.Value(contextKey{})
	if ctxValue == nil {
		return make(map[string]any)
	}
	ctxLabels, ok := ctxValue.(map[string]any)
	if !ok {
		return make(map[string]any)
	}

	return maps.Clone(ctxLabels)
}
