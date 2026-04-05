package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

const (
	StreamName = "GONDOR_EVENTS"
	// SubjectPrefix is the base subject for all Gondor events.
	SubjectPrefix = "gondor.events"
)

// Client wraps a NATS connection and JetStream context.
type Client struct {
	conn   *nats.Conn
	js     jetstream.JetStream
	logger *zap.Logger
}

// Connect establishes a NATS connection and initialises JetStream.
func Connect(url string, logger *zap.Logger) (*Client, error) {
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("jetstream init: %w", err)
	}

	client := &Client{conn: nc, js: js, logger: logger}

	// Ensure the stream exists
	if err := client.ensureStream(); err != nil {
		nc.Close()
		return nil, err
	}

	logger.Info("connected to NATS JetStream", zap.String("url", url))
	return client, nil
}

func (c *Client) ensureStream() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      StreamName,
		Subjects:  []string{SubjectPrefix + ".>"},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour, // keep events for 7 days
		Storage:   jetstream.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("ensure stream: %w", err)
	}
	c.logger.Info("JetStream stream ready", zap.String("stream", StreamName))
	return nil
}

// Publish sends a message to the given event type subject.
func (c *Client) Publish(ctx context.Context, eventType string, data []byte) error {
	subject := SubjectPrefix + "." + eventType
	_, err := c.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	c.logger.Debug("event published", zap.String("subject", subject))
	return nil
}

// Subscribe creates a durable consumer for the given event type pattern.
func (c *Client) Subscribe(ctx context.Context, consumerName string, subjectFilter string, handler func(msg jetstream.Msg)) (jetstream.ConsumeContext, error) {
	subject := SubjectPrefix + "." + subjectFilter

	consumer, err := c.js.CreateOrUpdateConsumer(ctx, StreamName, jetstream.ConsumerConfig{
		Durable:       consumerName,
		FilterSubject: subject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    5,
		AckWait:       30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("create consumer %s: %w", consumerName, err)
	}

	cc, err := consumer.Consume(handler)
	if err != nil {
		return nil, fmt.Errorf("consume %s: %w", consumerName, err)
	}

	c.logger.Info("subscribed", zap.String("consumer", consumerName), zap.String("subject", subject))
	return cc, nil
}

// Close drains and closes the NATS connection.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Drain()
	}
}
