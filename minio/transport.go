package minio

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	metricClientReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: "minio_client",
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "minio client requests duration(ms).",
		Labels:    []string{"method", "bucket", "endpoint"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500},
	})
	metricClientReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "minio_client",
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "minio client requests code count.",
		Labels:    []string{"method", "bucket", "code", "endpoint"},
	})
	metricClientFailoverTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "minio_client",
		Subsystem: "requests",
		Name:      "failover_total",
		Help:      "minio client failover count when primary node fails.",
		Labels:    []string{"endpoint"},
	})
	metricClientAffinityHitTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "minio_client",
		Subsystem: "affinity",
		Name:      "hit_total",
		Help:      "minio client affinity cache hit count.",
		Labels:    []string{"bucket"},
	})
	metricClientAffinityMissTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "minio_client",
		Subsystem: "affinity",
		Name:      "miss_total",
		Help:      "minio client affinity cache miss count.",
		Labels:    []string{"bucket"},
	})
	metricClientBreakerTripTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "minio_client",
		Subsystem: "breaker",
		Name:      "trip_total",
		Help:      "minio client circuit breaker trip count.",
		Labels:    []string{"endpoint"},
	})
)

const (
	instrumentationName  = "github.com/lerity-yao/czt-contrib/minio"
	clientSpanName       = "minio-client"
	metricCodeNoResponse = "0"
)

// instrumentedTransport wraps an http.RoundTripper with go-zero observability:
// tracing, circuit breaking, metrics, and logging.
type instrumentedTransport struct {
	base          http.RoundTripper
	brk           breaker.Breaker
	name          string
	slowThreshold int64
}

// newInstrumentedTransport creates an instrumented transport for the given endpoint.
func newInstrumentedTransport(endpoint string, base http.RoundTripper, slowThreshold int64) *instrumentedTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &instrumentedTransport{
		base:          base,
		brk:           breaker.NewBreaker(breaker.WithName("minio:" + endpoint)),
		name:          endpoint,
		slowThreshold: slowThreshold,
	}
}

// RoundTrip executes a single HTTP transaction with tracing, circuit breaking,
// metrics recording, and slow-request logging.
func (t *instrumentedTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	ctx := req.Context()

	// Start trace span.
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)
	ctx, span := tracer.Start(ctx, clientSpanName+"."+req.Method,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// Inject trace context into outgoing headers.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	req = req.WithContext(ctx)

	// Extract bucket name from path for metric labels.
	bucket := extractBucket(req.URL.Path)

	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		metricClientReqDur.Observe(duration, req.Method, bucket, t.name)

		code := metricCodeNoResponse
		if resp != nil {
			code = strconv.Itoa(resp.StatusCode)
		}
		metricClientReqCodeTotal.Inc(req.Method, bucket, code, t.name)

		// Log slow requests or errors. 0 means disable slow logging.
		if (t.slowThreshold > 0 && duration > t.slowThreshold) || err != nil {
			logx.WithContext(ctx).Slowf("[minio] %s %s bucket=%s duration=%dms err=%v",
				req.Method, req.URL.Path, bucket, duration, err)
		}
	}()

	// Execute under circuit breaker protection.
	err = t.brk.DoWithAcceptable(func() error {
		resp, err = t.base.RoundTrip(req)
		return err
	}, func(err error) bool {
		// Consider server errors (5xx) as failures for the breaker.
		return resp != nil && resp.StatusCode < 500
	})

	if errors.Is(err, breaker.ErrServiceUnavailable) {
		metricClientBreakerTripTotal.Inc(t.name)
	}

	return resp, err
}

// extractBucket extracts the bucket name from the URL path.
// MinIO paths are typically /{bucket}/{key...}.
func extractBucket(path string) string {
	path = strings.TrimPrefix(path, "/")
	if idx := strings.Index(path, "/"); idx > 0 {
		return path[:idx]
	}
	return path
}
