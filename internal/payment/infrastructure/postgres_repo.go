package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"example.com/myapp/payment/application"
	"example.com/myapp/payment/domain"
	"example.com/myapp/shared/db"
	"github.com/jackc/pgx/v4"
)

// PostgresPaymentRepository는 PostgreSQL을 사용하는 결제 저장소 구현체입니다.
type PostgresPaymentRepository struct {
	db *db.Database
}

// NewPostgresPaymentRepository는 새로운 PostgresPaymentRepository 인스턴스를 생성합니다.
func NewPostgresPaymentRepository(database *db.Database) application.PaymentRepository {
	return &PostgresPaymentRepository{
		db: database,
	}
}

// Save는 결제 정보를 데이터베이스에 저장합니다.
func (r *PostgresPaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	// 추가 결제 데이터를 JSON으로 변환
	paymentDataJSON, err := json.Marshal(payment.PaymentData())
	if err != nil {
		return fmt.Errorf("failed to marshal payment data: %w", err)
	}

	query := `
		INSERT INTO payments (id, order_id, amount, method, status, transaction_id, payment_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Pool.Exec(
		ctx,
		query,
		payment.ID(),
		payment.OrderID(),
		payment.Amount(),
		string(payment.Method()),
		string(payment.Status()),
		payment.TransactionID(),
		paymentDataJSON,
		payment.CreatedAt(),
		payment.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	return nil
}

// FindByID는 ID로 결제를 조회합니다.
func (r *PostgresPaymentRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, amount, method, status, transaction_id, payment_data, created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)

	var paymentID, orderID, methodStr, statusStr, transactionID string
	var amount float64
	var paymentDataJSON []byte
	var createdAt, updatedAt string

	err := row.Scan(
		&paymentID,
		&orderID,
		&amount,
		&methodStr,
		&statusStr,
		&transactionID,
		&paymentDataJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment by ID: %w", err)
	}

	// JSON에서 결제 데이터 파싱
	var paymentData map[string]string
	if err := json.Unmarshal(paymentDataJSON, &paymentData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment data: %w", err)
	}

	// 실제 구현에서는 DB 레코드를 도메인 엔티티로 변환하는 로직이 필요합니다.
	// 여기서는 코드 예시를 간략하게 하기 위해 생략합니다.
	return &domain.Payment{}, nil
}

// FindByOrderID는 주문 ID로 결제를 조회합니다.
func (r *PostgresPaymentRepository) FindByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, amount, method, status, transaction_id, payment_data, created_at, updated_at
		FROM payments
		WHERE order_id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, orderID)

	var paymentID, retrievedOrderID, methodStr, statusStr, transactionID string
	var amount float64
	var paymentDataJSON []byte
	var createdAt, updatedAt string

	err := row.Scan(
		&paymentID,
		&retrievedOrderID,
		&amount,
		&methodStr,
		&statusStr,
		&transactionID,
		&paymentDataJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment by order ID: %w", err)
	}

	// JSON에서 결제 데이터 파싱
	var paymentData map[string]string
	if err := json.Unmarshal(paymentDataJSON, &paymentData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment data: %w", err)
	}

	// 실제 구현에서는 DB 레코드를 도메인 엔티티로 변환하는 로직이 필요합니다.
	return &domain.Payment{}, nil
}

// Update는 결제 정보를 업데이트합니다.
func (r *PostgresPaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	// 추가 결제 데이터를 JSON으로 변환
	paymentDataJSON, err := json.Marshal(payment.PaymentData())
	if err != nil {
		return fmt.Errorf("failed to marshal payment data: %w", err)
	}

	query := `
		UPDATE payments
		SET status = $1, transaction_id = $2, payment_data = $3, updated_at = $4
		WHERE id = $5
	`

	_, err = r.db.Pool.Exec(
		ctx,
		query,
		string(payment.Status()),
		payment.TransactionID(),
		paymentDataJSON,
		payment.UpdatedAt(),
		payment.ID(),
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	return nil
}