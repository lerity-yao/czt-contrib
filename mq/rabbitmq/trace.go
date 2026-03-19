package rabbitmq

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "github.com/lerity-yao/czt-contrib/mq/rabbitmq"
	producerSpanName    = "rabbitmq-producer"
	consumerSpanName    = "rabbitmq-consumer"
)

// StartProducerSpan 开启生产者 Span (Sender 端使用)
func StartProducerSpan(ctx context.Context, exchange string, routeKey string) (context.Context, oteltrace.Span) {
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)

	// 标记为 Producer 种类
	childCtx, span := tracer.Start(ctx,
		fmt.Sprintf("%s-%s", producerSpanName, routeKey),
		oteltrace.WithSpanKind(oteltrace.SpanKindProducer),
	)

	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", exchange),
		attribute.String("messaging.operation", "send"),
	)

	return childCtx, span
}

// StartConsumerSpan 开启消费者 Span (Listener 中间件使用)
func StartConsumerSpan(ctx context.Context, queueName string, carrier *propagation.HeaderCarrier) (context.Context, oteltrace.Span) {
	// 1. 提取上游 trace 上下文（carrier 为 nil 时使用原始 ctx）
	extractedCtx := ctx
	if carrier != nil {
		extractedCtx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}

	// 2. 开启消费者 Span
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)
	childCtx, span := tracer.Start(extractedCtx,
		fmt.Sprintf("%s-%s", consumerSpanName, queueName),
		oteltrace.WithSpanKind(oteltrace.SpanKindConsumer),
	)

	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", queueName),
		attribute.String("messaging.operation", "process"),
	)

	return childCtx, span
}

// EndSpan 统一结束 Span 的逻辑（完全对齐 go-zero sqlx 风格）
func EndSpan(span oteltrace.Span, err error) {
	defer span.End()

	if err == nil {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}
