package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"flipay/internal/payment"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Sender struct {
	url    string
	secret string
	client *http.Client
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewSender(url, secret string, db *pgxpool.Pool, logger *zap.Logger) *Sender {
	return &Sender{url: url, secret: secret, db: db, logger: logger, client: &http.Client{Timeout: 5 * time.Second}}
}

// Send simulates merchant webhook delivery. Failed deliveries are recorded for future retry work.
func (s *Sender) Send(ctx context.Context, p payment.Payment) error {
	payload, _ := json.Marshal(map[string]any{
		"payment_id":   p.ID,
		"reference_no": p.ReferenceNo,
		"status":       p.Status,
		"amount":       p.Amount,
		"occurred_at":  time.Now().UTC(),
	})

	status := "FAILED"
	err := s.deliver(ctx, payload)
	if err == nil {
		status = "SUCCESS"
	}

	_, dbErr := s.db.Exec(ctx, `INSERT INTO callbacks (payment_id, payload, status, attempts, last_error) VALUES ($1,$2,$3,$4,$5)`, p.ID, string(payload), status, 1, errorString(err))
	if dbErr != nil {
		s.logger.Error("store webhook callback failed", zap.String("payment_id", p.ID), zap.Error(dbErr))
	}
	if err != nil {
		return err
	}
	s.logger.Info("webhook delivered", zap.String("payment_id", p.ID), zap.String("status", string(p.Status)))
	return nil
}

func (s *Sender) deliver(ctx context.Context, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Flipay-Signature", sign(payload, s.secret))

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", res.StatusCode)
	}
	return nil
}

func ValidateSignature(payload []byte, secret, signature string) bool {
	return hmac.Equal([]byte(sign(payload, secret)), []byte(signature))
}

func sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
