package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository isolates database access from business rules.
type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, user User) (User, error) {
	// A transaction keeps the write path ready for future auth-side effects,
	// such as audit logs, profile rows, or email verification tokens.
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO users (id, name, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`
	err = tx.QueryRow(ctx, query, user.ID, user.Name, user.Email, user.PasswordHash).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrEmailAlreadyExists
		}
		return User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	return r.findOne(ctx, `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email=$1`, email)
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (User, error) {
	return r.findOne(ctx, `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id=$1`, id)
}

func (r *PostgresRepository) findOne(ctx context.Context, query string, arg string) (User, error) {
	var user User
	err := r.db.QueryRow(ctx, query, arg).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	return user, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
