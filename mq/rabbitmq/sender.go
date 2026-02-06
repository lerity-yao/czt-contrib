package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func MustNewSender(rabbitMqConf RabbitSenderConf) (Sender, error) {
	sender := &RabbitMqSender{ContentType: rabbitMqConf.ContentType}
	conn, err := amqp.Dial(getRabbitURL(rabbitMqConf.RabbitConf))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect rabbitmq, error: %v", err))
	}

	sender.conn = conn
	channel, err := sender.conn.Channel()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to open a channel, error: %v", err))
	}
	sender.channel = channel
	return sender, nil
}

func (q *RabbitMqSender) Send(ctx context.Context, exchange string, routeKey string, msg []byte) error {

	spanName := fmt.Sprintf("%s-producer", routeKey)

	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)

	if span := oteltrace.SpanFromContext(ctx); !span.SpanContext().IsValid() {

		opts := []sdktrace.TracerProviderOption{
			// Set the sampling rate based on the parent span to 100%
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))),
			// Record information about this application in a Resource.
			sdktrace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(spanName))),
		}
		tp := sdktrace.NewTracerProvider(opts...)
		otel.SetTracerProvider(tp)
		tracer = tp.Tracer(trace.TraceName)
		ctx, _ = tracer.Start(ctx, spanName, oteltrace.WithSpanKind(oteltrace.SpanKindProducer))
		defer tp.Shutdown(ctx)
	}

	injectSpanCtx, injectSpan := tracer.Start(ctx, spanName, oteltrace.WithSpanKind(oteltrace.SpanKindProducer))
	carrier := &propagation.HeaderCarrier{}
	otel.GetTextMapPropagator().Inject(injectSpanCtx, carrier)

	msgBody := &RabbitMsgBody{
		Carrier: carrier,
		Msg:     msg,
	}

	msgBodyMap, err := json.Marshal(msgBody)

	if err != nil {
		return err
	}

	err = q.channel.PublishWithContext(
		injectSpanCtx,
		exchange,
		routeKey,
		false,
		false,
		amqp.Publishing{
			ContentType: q.ContentType,
			Body:        msgBodyMap,
		},
	)
	defer injectSpan.End()
	if err != nil {
		logc.Infof(ctx, "Failed to publish a message, error: %v", err)
		return err
	}
	logc.Infof(ctx, "Successfully publish a message, message: %v", string(msg))
	return nil
}

func (q *RabbitMqSender) Close() error {
	q.channel.Close()
	q.conn.Close()
	return nil
}
