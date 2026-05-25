package payment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"flipay/internal/queue"
	"flipay/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service owns payment business rules and orchestrates repository + queue.
type Service interface {
	Create(ctx context.Context, userID, idempotencyKey string, req CreatePaymentRequest) (CreatePaymentResponse, error)
	FindByID(ctx context.Context, id, userID string) (PaymentResponse, error)
	History(ctx context.Context, userID string) ([]PaymentResponse, error)
}

type service struct {
	repo          Repository
	queue         queue.Queue
	expiryMinutes int
	logger        *zap.Logger
}

func NewService(repo Repository, queue queue.Queue, expiryMinutes int, logger *zap.Logger) Service {
	return &service{repo: repo, queue: queue, expiryMinutes: expiryMinutes, logger: logger}
}

func (s *service) Create(ctx context.Context, userID, idempotencyKey string, req CreatePaymentRequest) (CreatePaymentResponse, error) {
	requestHash := hashRequest(req)
	if idempotencyKey != "" {
		cachedHash, cachedResponse, err := s.repo.FindIdempotency(ctx, idempotencyKey)
		if err == nil {
			if cachedHash != requestHash {
				return CreatePaymentResponse{}, ErrIdempotencyConflict
			}
			var replay CreatePaymentResponse
			if json.Unmarshal([]byte(cachedResponse), &replay) == nil {
				replay.Idempotent = true
				return replay, nil
			}
		} else if !errors.Is(err, ErrPaymentNotFound) {
			return CreatePaymentResponse{}, err
		}
	}

	p := Payment{
		ID:            uuid.NewString(),
		UserID:        userID,
		ReferenceNo:   utils.GenerateReferenceNo(),
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Status:        StatusPending,
		ExpiredAt:     time.Now().Add(time.Duration(s.expiryMinutes) * time.Minute),
	}
	if req.PaymentMethod == MethodBankTransfer {
		p.VANumber = utils.GenerateVANumber()
	}
	if req.PaymentMethod == MethodQRIS {
		p.QRISString = utils.GenerateQRISString(p.ReferenceNo, req.Amount)
	}

	response := CreatePaymentResponse{
		Payment:     NewPaymentResponse(p),
		Queued:      true,
		StatusCheck: "/api/v1/payments/" + p.ID,
	}
	responsePayload, _ := json.Marshal(response)

	created, err := s.repo.CreateWithIdempotency(ctx, p, idempotencyKey, requestHash, string(responsePayload))
	if err != nil {
		return CreatePaymentResponse{}, err
	}
	response.Payment = NewPaymentResponse(created)

	// Queue push happens after DB commit. If Redis is down, the payment stays PENDING and can be retried later.
	if err := s.queue.PushPayment(ctx, queue.PaymentJob{PaymentID: created.ID, Attempt: 1}); err != nil {
		s.logger.Error("push payment queue failed", zap.String("payment_id", created.ID), zap.Error(err))
		response.Queued = false
		return response, nil
	}

	s.logger.Info("payment created", zap.String("payment_id", created.ID), zap.String("reference_no", created.ReferenceNo))
	return response, nil
}

func (s *service) FindByID(ctx context.Context, id, userID string) (PaymentResponse, error) {
	p, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return PaymentResponse{}, err
	}
	return NewPaymentResponse(p), nil
}

func (s *service) History(ctx context.Context, userID string) ([]PaymentResponse, error) {
	payments, err := s.repo.FindHistory(ctx, userID)
	if err != nil {
		return nil, err
	}
	responses := make([]PaymentResponse, 0, len(payments))
	for _, p := range payments {
		responses = append(responses, NewPaymentResponse(p))
	}
	return responses, nil
}

func hashRequest(req CreatePaymentRequest) string {
	payload, _ := json.Marshal(req)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
