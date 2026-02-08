package cron

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "github.com/lerity-yao/czt-contrib/cron"
	producerSpanName    = "asynq-producer"
	consumerSpanName    = "asynq-consumer"
)

// StartProducerSpan 开启生产者 Span (Client 端使用)
func StartProducerSpan(ctx context.Context, taskType string) (context.Context, oteltrace.Span) {
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)

	// 标记为 Producer 种类
	childCtx, span := tracer.Start(ctx,
		fmt.Sprintf("%s-%s", producerSpanName, taskType),
		oteltrace.WithSpanKind(oteltrace.SpanKindProducer),
	)

	span.SetAttributes(
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.destination", taskType),
		attribute.String("messaging.operation", "send"),
	)

	return childCtx, span
}

// StartConsumerSpan 开启消费者 Span (Server 中间件使用)
func StartConsumerSpan(ctx context.Context, t *asynq.Task) (context.Context, oteltrace.Span) {

	header := t.Headers()
	extractedCtx := otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(header))

	// 2. 开启消费者 Span
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)
	childCtx, span := tracer.Start(extractedCtx,
		fmt.Sprintf("%s-%s", consumerSpanName, t.Type()),
		oteltrace.WithSpanKind(oteltrace.SpanKindConsumer),
	)

	span.SetAttributes(
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.destination", t.Type()),
		attribute.String("messaging.operation", "process"),
	)

	return childCtx, span
}

// EndSpan 统一结束 Span 的逻辑 (完全对齐 go-zero sqlx 风格)
func EndSpan(span oteltrace.Span, err error) {
	defer span.End()

	if err == nil {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}
