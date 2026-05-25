package payment

import "time"

// CreatePaymentRequest is validated by Gin before reaching the service layer.
type CreatePaymentRequest struct {
	Amount        int64  `json:"amount" binding:"required,min=1000"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=bank_transfer qris"`
}

// PaymentResponse is the public API response. Internal fields stay inside the domain model.
type PaymentResponse struct {
	ID            string    `json:"id"`
	ReferenceNo   string    `json:"reference_no"`
	Amount        int64     `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	VANumber      string    `json:"va_number,omitempty"`
	QRISString    string    `json:"qris_string,omitempty"`
	Status        Status    `json:"status"`
	ExpiredAt     time.Time `json:"expired_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreatePaymentResponse struct {
	Payment     PaymentResponse `json:"payment"`
	Idempotent  bool            `json:"idempotent"`
	Queued      bool            `json:"queued"`
	StatusCheck string          `json:"status_check"`
}

func NewPaymentResponse(payment Payment) PaymentResponse {
	return PaymentResponse{
		ID:            payment.ID,
		ReferenceNo:   payment.ReferenceNo,
		Amount:        payment.Amount,
		PaymentMethod: payment.PaymentMethod,
		VANumber:      payment.VANumber,
		QRISString:    payment.QRISString,
		Status:        payment.Status,
		ExpiredAt:     payment.ExpiredAt,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}
}
