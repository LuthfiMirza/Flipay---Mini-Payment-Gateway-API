package worker

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"flipay/internal/payment"
	"flipay/internal/queue"
	"flipay/internal/webhook"
	"go.uber.org/zap"
)

const maxPaymentAttempts = 3

type PaymentWorker struct {
	repo    payment.Repository
	queue   queue.Queue
	webhook *webhook.Sender
	logger  *zap.Logger
}

func NewPaymentWorker(repo payment.Repository, queue queue.Queue, webhook *webhook.Sender, logger *zap.Logger) *PaymentWorker {
	return &PaymentWorker{repo: repo, queue: queue, webhook: webhook, logger: logger}
}

func (w *PaymentWorker) Start(ctx context.Context) {
	go w.expirationLoop(ctx)
	go w.paymentLoop(ctx)
}

func (w *PaymentWorker) paymentLoop(ctx context.Context) {
	for {
		job, err := w.queue.PopPayment(ctx)
		if err != nil {
			w.logger.Error("pop payment job failed", zap.Error(err))
			continue
		}
		w.processPayment(ctx, job)
	}
}

func (w *PaymentWorker) processPayment(ctx context.Context, job queue.PaymentJob) {
	if job.Attempt == 0 {
		job.Attempt = 1
	}

	// Simulate external payment processor latency.
	time.Sleep(2 * time.Second)

	status := payment.StatusSuccess
	if rand.Intn(10) == 0 {
		status = payment.StatusFailed
	}

	updatedPayment, err := w.repo.UpdateStatus(ctx, job.PaymentID, []payment.Status{payment.StatusPending}, status)
	if err != nil {
		if errors.Is(err, payment.ErrInvalidPaymentStatus) {
			w.logger.Info("payment skipped because status already changed", zap.String("payment_id", job.PaymentID))
			return
		}
		w.retry(ctx, job, err)
		return
	}

	if err := w.webhook.Send(ctx, updatedPayment); err != nil {
		w.logger.Error("webhook send failed", zap.String("payment_id", job.PaymentID), zap.Error(err))
	}
	w.logger.Info("payment processed", zap.String("payment_id", job.PaymentID), zap.String("status", string(updatedPayment.Status)))
}

func (w *PaymentWorker) retry(ctx context.Context, job queue.PaymentJob, cause error) {
	if job.Attempt >= maxPaymentAttempts {
		w.logger.Error("payment job exhausted retries", zap.String("payment_id", job.PaymentID), zap.Int("attempt", job.Attempt), zap.Error(cause))
		return
	}
	job.Attempt++
	delay := time.Duration(job.Attempt) * 2 * time.Second
	w.logger.Warn("retrying payment job", zap.String("payment_id", job.PaymentID), zap.Int("attempt", job.Attempt), zap.Duration("delay", delay), zap.Error(cause))
	_ = w.queue.PushPaymentWithDelay(ctx, job, delay)
}

func (w *PaymentWorker) expirationLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expired, err := w.repo.ExpirePending(ctx)
			if err != nil {
				w.logger.Error("expire pending payments failed", zap.Error(err))
				continue
			}
			if expired > 0 {
				w.logger.Info("payments expired", zap.Int64("count", expired))
			}
		}
	}
}
