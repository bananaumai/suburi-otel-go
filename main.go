package main

import (
	"context"
	"flag"
	"go.opentelemetry.io/otel"
	"io"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

func newExporter(w io.Writer) (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("fib"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	return r
}

func main() {
	var samplingFraction float64
	flag.Float64Var(&samplingFraction, "sampling-fraction", 1, "sampling fraction")
	flag.Parse()

	exp, err := newExporter(os.Stderr)
	if err != nil {
		panic(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newResource()),
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(samplingFraction))),
	)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()
	otel.SetTracerProvider(tp)

	wg := &sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			f0()
		}()
	}
	wg.Wait()
}

func f0() {
	ctx, span := otel.Tracer("").Start(context.Background(), "f0")
	defer span.End()
	span.SetName("root")
	for i := 0; i < 10; i++ {
		f1(ctx)
	}
}

func f1(ctx context.Context) {
	ctx, span := otel.Tracer("").Start(context.Background(), "f1")
	defer span.End()
	f2(ctx)
	time.Sleep(time.Millisecond)
}

func f2(ctx context.Context) {
	ctx, span := otel.Tracer("").Start(context.Background(), "f2")
	defer span.End()
	time.Sleep(time.Millisecond)
}
