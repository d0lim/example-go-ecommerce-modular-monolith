package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"example.com/myapp/order/application"
	"example.com/myapp/order/domain"
	"example.com/myapp/shared/db"
	"github.com/jackc/pgx/v4"
)

// PostgresOrderRepository는 PostgreSQL을 사용하는 주문 저장소 구현체입니다.
type PostgresOrderRepository struct {
	db *db.Database
}

// NewPostgresOrderRepository는 새로운 PostgresOrderRepository 인스턴스를 생성합니다.
func NewPostgresOrderRepository(database *db.Database) application.OrderRepository {
	return &PostgresOrderRepository{
		db: database,
	}
}

// Save는 주문 정보를 데이터베이스에 저장합니다.
func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // 실패 시 트랜잭션 롤백

	// 1. 주문 기본 정보 저장
	orderQuery := `
		INSERT INTO orders (id, customer_id, total_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.Exec(
		ctx,
		orderQuery,
		order.ID(),
		order.CustomerID(),
		order.TotalAmount(),
		string(order.Status()),
		order.CreatedAt(),
		order.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// 2. 주문 항목 저장
	for _, item := range order.Items() {
		itemQuery := `
			INSERT INTO order_items (id, order_id, product_id, name, price, quantity)
			VALUES ($1, $2, $3, $4, $5, $6)
		`

		_, err = tx.Exec(
			ctx,
			itemQuery,
			item.ID(),
			order.ID(),
			item.ProductID(),
			item.Name(),
			item.Price(),
			item.Quantity(),
		)

		if err != nil {
			return fmt.Errorf("failed to save order item: %w", err)
		}
	}

	// 트랜잭션 커밋
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID는 ID로 주문을 조회합니다.
func (r *PostgresOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	// 1. 주문 기본 정보 조회
	orderQuery := `
		SELECT id, customer_id, total_amount, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, orderQuery, id)

	var orderID, customerID, status string
	var totalAmount float64
	var createdAt, updatedAt string

	err := row.Scan(&orderID, &customerID, &totalAmount, &status, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to find order by ID: %w", err)
	}

	// 2. 주문 항목 조회
	itemsQuery := `
		SELECT id, product_id, name, price, quantity
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.Pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	items := []*domain.OrderItem{}
	for rows.Next() {
		var itemID, productID, name string
		var price float64
		var quantity int

		if err := rows.Scan(&itemID, &productID, &name, &price, &quantity); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		item := domain.NewOrderItem(productID, name, price, quantity)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	// 실제 구현에서는 DB 레코드를 도메인 엔티티로 복원하는 로직이 필요합니다.
	// 여기서는 코드 예시를 간략하게 하기 위해 생략합니다.
	return &domain.Order{}, nil
}

// FindByCustomerID는 고객 ID로 주문 목록을 조회합니다.
func (r *PostgresOrderRepository) FindByCustomerID(ctx context.Context, customerID string) ([]*domain.Order, error) {
	query := `
		SELECT id
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders by customer ID: %w", err)
	}
	defer rows.Close()

	orderIDs := []string{}
	for rows.Next() {
		var orderID string
		if err := rows.Scan(&orderID); err != nil {
			return nil, fmt.Errorf("failed to scan order ID: %w", err)
		}
		orderIDs = append(orderIDs, orderID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order IDs: %w", err)
	}

	// 주문 ID별로 상세 정보 조회
	orders := []*domain.Order{}
	for _, orderID := range orderIDs {
		order, err := r.FindByID(ctx, orderID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// Update는 주문 정보를 업데이트합니다.
func (r *PostgresOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		string(order.Status()),
		order.UpdatedAt(),
		order.ID(),
	)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// Delete는 주문을 삭제합니다.
func (r *PostgresOrderRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // 실패 시 트랜잭션 롤백

	// 1. 주문 항목 삭제
	_, err = tx.Exec(ctx, "DELETE FROM order_items WHERE order_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	// 2. 주문 삭제
	result, err := tx.Exec(ctx, "DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	// 영향받은 행이 없으면 주문이 존재하지 않음
	if result.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}

	// 트랜잭션 커밋
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}