package observability

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const superviseRestartBackoff = time.Second

// SuperviseLoop runs fn in a panic-safe loop until ctx is cancelled.
func SuperviseLoop(ctx context.Context, name string, fn func(context.Context)) {
	for {
		if ctx.Err() != nil {
			log.Infof("supervised loop %q exiting: %v", name, ctx.Err())
			return
		}
		runOnce(ctx, name, fn)
		select {
		case <-ctx.Done():
			log.Infof("supervised loop %q exiting: %v", name, ctx.Err())
			return
		case <-time.After(superviseRestartBackoff):
		}
	}
}

func runOnce(ctx context.Context, name string, fn func(context.Context)) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("supervised loop %q panicked: %v\n%s", name, r, debug.Stack())
			if m := GetMetrics(); m != nil && m.BgJobPanicsTotal != nil {
				m.BgJobPanicsTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("job", name)))
			}
		}
	}()
	fn(ctx)
}
