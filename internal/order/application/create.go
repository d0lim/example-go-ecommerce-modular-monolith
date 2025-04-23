package application

import (
	"context"
	"errors"

	"example.com/myapp/order/domain"
)

var (
	ErrInvalidCustomerID = errors.New("invalid customer ID")
	ErrOrderNotFound     = errors.New("order not found")
)

// CreateOrder는 새로운 주문을 생성합니다.
func (uc *OrderUseCase) CreateOrder(ctx context.Context, customerID string, itemRequests []OrderItemRequest) (*domain.Order, error) {
	if customerID == "" {
		return nil, ErrInvalidCustomerID
	}

	if len(itemRequests) == 0 {
		return nil, domain.ErrInvalidOrderItems
	}

	// OrderItemRequest를 도메인 OrderItem으로 변환
	items := make([]*domain.OrderItem, 0, len(itemRequests))
	for _, req := range itemRequests {
		item := domain.NewOrderItem(req.ProductID, req.Name, req.Price, req.Quantity)
		items = append(items, item)
	}

	// 새로운 주문 생성
	order, err := domain.NewOrder(customerID, items)
	if err != nil {
		return nil, err
	}

	// 저장소에 주문 저장
	if err := uc.repo.Save(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// GetOrder는 주문 ID로 주문을 조회합니다.
func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return uc.repo.FindByID(ctx, id)
}

// GetCustomerOrders는 고객 ID로 주문 목록을 조회합니다.
func (uc *OrderUseCase) GetCustomerOrders(ctx context.Context, customerID string) ([]*domain.Order, error) {
	if customerID == "" {
		return nil, ErrInvalidCustomerID
	}
	return uc.repo.FindByCustomerID(ctx, customerID)
}

// UpdateOrderStatus는 주문 상태를 업데이트합니다.
func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) (*domain.Order, error) {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := order.UpdateStatus(status); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// CancelOrder는 주문을 취소합니다.
func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	return uc.UpdateOrderStatus(ctx, id, domain.StatusCanceled)
}