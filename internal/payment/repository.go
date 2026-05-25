package payment

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository hides PostgreSQL details from the service layer.
type Repository interface {
	CreateWithIdempotency(ctx context.Context, payment Payment, idempotencyKey, requestHash, response string) (Payment, error)
	FindByID(ctx context.Context, id, userID string) (Payment, error)
	FindByReference(ctx context.Context, referenceNo string) (Payment, error)
	FindHistory(ctx context.Context, userID string) ([]Payment, error)
	FindIdempotency(ctx context.Context, key string) (requestHash string, response string, err error)
	UpdateStatus(ctx context.Context, id string, from []Status, to Status) (Payment, error)
	ExpirePending(ctx context.Context) (int64, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateWithIdempotency(ctx context.Context, p Payment, idempotencyKey, requestHash, response string) (Payment, error) {
	// Payment creation and idempotency storage must succeed or fail together.
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Payment{}, err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO payments (id, user_id, reference_no, amount, payment_method, va_number, qris_string, status, expired_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING created_at, updated_at`
	err = tx.QueryRow(ctx, query, p.ID, p.UserID, p.ReferenceNo, p.Amount, p.PaymentMethod, p.VANumber, p.QRISString, p.Status, p.ExpiredAt).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return Payment{}, err
	}

	if idempotencyKey != "" {
		_, err = tx.Exec(ctx, `INSERT INTO idempotencies (idempotency_key, request_hash, response) VALUES ($1,$2,$3)`, idempotencyKey, requestHash, response)
		if err != nil {
			return Payment{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Payment{}, err
	}
	return p, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id, userID string) (Payment, error) {
	if userID == "" {
		return r.findOne(ctx, selectPaymentSQL()+` WHERE id=$1`, id)
	}
	return r.findOne(ctx, selectPaymentSQL()+` WHERE id=$1 AND user_id=$2`, id, userID)
}

func (r *PostgresRepository) FindByReference(ctx context.Context, referenceNo string) (Payment, error) {
	return r.findOne(ctx, selectPaymentSQL()+` WHERE reference_no=$1`, referenceNo)
}

func (r *PostgresRepository) FindHistory(ctx context.Context, userID string) ([]Payment, error) {
	rows, err := r.db.Query(ctx, selectPaymentSQL()+` WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := make([]Payment, 0)
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

func (r *PostgresRepository) FindIdempotency(ctx context.Context, key string) (string, string, error) {
	var requestHash string
	var response string
	err := r.db.QueryRow(ctx, `SELECT request_hash, response FROM idempotencies WHERE idempotency_key=$1`, key).Scan(&requestHash, &response)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrPaymentNotFound
	}
	return requestHash, response, err
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id string, from []Status, to Status) (Payment, error) {
	query := selectPaymentSQL() + ` WHERE id=$1 AND status = ANY($2::text[]) FOR UPDATE`
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Payment{}, err
	}
	defer tx.Rollback(ctx)

	fromValues := make([]string, 0, len(from))
	for _, status := range from {
		fromValues = append(fromValues, string(status))
	}

	var p Payment
	err = tx.QueryRow(ctx, query, id, fromValues).Scan(&p.ID, &p.UserID, &p.ReferenceNo, &p.Amount, &p.PaymentMethod, &p.VANumber, &p.QRISString, &p.Status, &p.ExpiredAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, ErrInvalidPaymentStatus
	}
	if err != nil {
		return Payment{}, err
	}

	err = tx.QueryRow(ctx, `UPDATE payments SET status=$2, updated_at=NOW() WHERE id=$1 RETURNING updated_at`, id, to).Scan(&p.UpdatedAt)
	if err != nil {
		return Payment{}, err
	}
	p.Status = to

	if err := tx.Commit(ctx); err != nil {
		return Payment{}, err
	}
	return p, nil
}

func (r *PostgresRepository) ExpirePending(ctx context.Context) (int64, error) {
	commandTag, err := r.db.Exec(ctx, `UPDATE payments SET status=$1, updated_at=NOW() WHERE status=$2 AND expired_at < NOW()`, StatusExpired, StatusPending)
	return commandTag.RowsAffected(), err
}

func (r *PostgresRepository) findOne(ctx context.Context, query string, args ...any) (Payment, error) {
	p, err := scanPayment(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, ErrPaymentNotFound
	}
	return p, err
}

func selectPaymentSQL() string {
	return `SELECT id,user_id,reference_no,amount,payment_method,va_number,qris_string,status,expired_at,created_at,updated_at FROM payments`
}

type paymentScanner interface {
	Scan(dest ...any) error
}

func scanPayment(scanner paymentScanner) (Payment, error) {
	var p Payment
	err := scanner.Scan(&p.ID, &p.UserID, &p.ReferenceNo, &p.Amount, &p.PaymentMethod, &p.VANumber, &p.QRISString, &p.Status, &p.ExpiredAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}
