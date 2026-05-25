package payment

import "time"

// Status is kept as a string type so it is easy to store in PostgreSQL and return as JSON.
type Status string

const (
	StatusPending Status = "PENDING"
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
	StatusExpired Status = "EXPIRED"
)

const (
	MethodBankTransfer = "bank_transfer"
	MethodQRIS         = "qris"
)

// Payment represents the payments table and the core payment entity.
type Payment struct {
	ID            string
	UserID        string
	ReferenceNo   string
	Amount        int64
	PaymentMethod string
	VANumber      string
	QRISString    string
	Status        Status
	ExpiredAt     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
