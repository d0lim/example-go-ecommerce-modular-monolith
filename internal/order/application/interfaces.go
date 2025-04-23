package application

import (
	"context"

	"example.com/myapp/order/domain"
)

// OrderRepository는 주문 관련 영속성 인터페이스를 정의합니다.
type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	FindByID(ctx context.Context, id string) (*domain.Order, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, id string) error
}

// OrderService는 주문 관련 비즈니스 로직을 정의합니다.
type OrderService interface {
	CreateOrder(ctx context.Context, customerID string, items []OrderItemRequest) (*domain.Order, error)
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	GetCustomerOrders(ctx context.Context, customerID string) ([]*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) (*domain.Order, error)
	CancelOrder(ctx context.Context, id string) (*domain.Order, error)
}

// OrderItemRequest는 주문 항목 생성 요청 정보를 정의합니다.
type OrderItemRequest struct {
	ProductID string
	Name      string
	Price     float64
	Quantity  int
}

// OrderUseCase는 OrderService 구현체를 정의합니다.
type OrderUseCase struct {
	repo OrderRepository
}

// NewOrderUseCase는 새로운 OrderUseCase 인스턴스를 생성합니다.
func NewOrderUseCase(repo OrderRepository) *OrderUseCase {
	return &OrderUseCase{
		repo: repo,
	}
}