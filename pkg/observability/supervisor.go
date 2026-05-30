package observability

import (
	"context"
	"runtime/debug"

	"github.com/gofiber/fiber/v3/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// SuperviseLoop runs fn in a panic-safe loop until ctx is cancelled. fn is
// expected to be a long-running daemon that returns only when ctx is done; if
// it returns earlier or panics, the supervisor restarts it. Every panic is
// logged with a stack trace and emits the bg_job_panics_total counter so
// silent crashes show up in alerting instead of just stopping the loop.
func SuperviseLoop(ctx context.Context, name string, fn func(context.Context)) {
	for {
		if ctx.Err() != nil {
			log.Infof("supervised loop %q exiting: %v", name, ctx.Err())
			return
		}
		runOnce(ctx, name, fn)
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
