package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	PaymentQueue           = "flipay:payments"
	PaymentDeadLetterQueue = "flipay:payments:dead"
)

// PaymentJob is the small payload sent to Redis. Keep queue messages compact.
type PaymentJob struct {
	PaymentID string `json:"payment_id"`
	Attempt   int    `json:"attempt"`
}

type Queue interface {
	PushPayment(ctx context.Context, job PaymentJob) error
	PushPaymentWithDelay(ctx context.Context, job PaymentJob, delay time.Duration) error
	PushDeadLetter(ctx context.Context, job PaymentJob, reason string) error
	PopPayment(ctx context.Context) (PaymentJob, error)
}

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(client *redis.Client) Queue {
	return &RedisQueue{client: client}
}

func (q *RedisQueue) PushPayment(ctx context.Context, job PaymentJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, PaymentQueue, payload).Err()
}

// PushPaymentWithDelay is a retry skeleton. For a real production queue, use a sorted set or stream.
func (q *RedisQueue) PushPaymentWithDelay(ctx context.Context, job PaymentJob, delay time.Duration) error {
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			_ = q.PushPayment(context.Background(), job)
		}
	}()
	return nil
}

func (q *RedisQueue) PushDeadLetter(ctx context.Context, job PaymentJob, reason string) error {
	payload, err := json.Marshal(map[string]any{
		"job":       job,
		"reason":    reason,
		"failed_at": time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, PaymentDeadLetterQueue, payload).Err()
}

func (q *RedisQueue) PopPayment(ctx context.Context) (PaymentJob, error) {
	result, err := q.client.BLPop(ctx, 0, PaymentQueue).Result()
	if err != nil {
		return PaymentJob{}, err
	}
	var job PaymentJob
	return job, json.Unmarshal([]byte(result[1]), &job)
}
