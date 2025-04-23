package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// OrderStatus는 주문 상태를 정의합니다.
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCanceled  OrderStatus = "canceled"
)

var (
	ErrInvalidOrderAmount   = errors.New("invalid order amount")
	ErrInvalidOrderItems    = errors.New("order must have at least one item")
	ErrInvalidOrderStatus   = errors.New("invalid order status")
	ErrOrderNotFound        = errors.New("order not found")
	ErrOrderStatusTransition = errors.New("invalid order status transition")
)

// OrderItem은 주문 항목을 나타냅니다.
type OrderItem struct {
	id        string
	productID string
	name      string
	price     float64
	quantity  int
}

// NewOrderItem은 새로운 주문 항목을 생성합니다.
func NewOrderItem(productID, name string, price float64, quantity int) *OrderItem {
	return &OrderItem{
		id:        uuid.New().String(),
		productID: productID,
		name:      name,
		price:     price,
		quantity:  quantity,
	}
}

// ID는 주문 항목의 고유 식별자를 반환합니다.
func (i *OrderItem) ID() string {
	return i.id
}

// ProductID는 상품 ID를 반환합니다.
func (i *OrderItem) ProductID() string {
	return i.productID
}

// Name은 상품 이름을 반환합니다.
func (i *OrderItem) Name() string {
	return i.name
}

// Price는 상품 단가를 반환합니다.
func (i *OrderItem) Price() float64 {
	return i.price
}

// Quantity는 주문 수량을 반환합니다.
func (i *OrderItem) Quantity() int {
	return i.quantity
}

// Subtotal은 상품별 소계를 반환합니다.
func (i *OrderItem) Subtotal() float64 {
	return i.price * float64(i.quantity)
}

// Order는 주문 엔티티를 나타냅니다.
type Order struct {
	id         string
	customerID string
	items      []*OrderItem
	totalAmount float64
	status     OrderStatus
	createdAt  time.Time
	updatedAt  time.Time
}

// NewOrder는 새로운 주문을 생성합니다.
func NewOrder(customerID string, items []*OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrInvalidOrderItems
	}

	// 총 금액 계산
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Subtotal()
	}

	if totalAmount <= 0 {
		return nil, ErrInvalidOrderAmount
	}

	now := time.Now()
	return &Order{
		id:          uuid.New().String(),
		customerID:  customerID,
		items:       items,
		totalAmount: totalAmount,
		status:      StatusPending,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ID는 주문의 고유 식별자를 반환합니다.
func (o *Order) ID() string {
	return o.id
}

// CustomerID는 고객 ID를 반환합니다.
func (o *Order) CustomerID() string {
	return o.customerID
}

// Items는 주문 항목 목록을 반환합니다.
func (o *Order) Items() []*OrderItem {
	return o.items
}

// TotalAmount는 주문 총액을 반환합니다.
func (o *Order) TotalAmount() float64 {
	return o.totalAmount
}

// Status는 주문 상태를 반환합니다.
func (o *Order) Status() OrderStatus {
	return o.status
}

// CreatedAt은 주문이 생성된 시간을 반환합니다.
func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

// UpdatedAt은 주문 정보가 마지막으로 업데이트된 시간을 반환합니다.
func (o *Order) UpdatedAt() time.Time {
	return o.updatedAt
}

// UpdateStatus는 주문 상태를 업데이트합니다.
func (o *Order) UpdateStatus(status OrderStatus) error {
	// 상태 전환 유효성 검사
	if !isValidStatusTransition(o.status, status) {
		return ErrOrderStatusTransition
	}

	o.status = status
	o.updatedAt = time.Now()
	return nil
}

// isValidStatusTransition은 주문 상태 전환이 유효한지 확인합니다.
func isValidStatusTransition(from, to OrderStatus) bool {
	// 상태 전환 규칙
	switch from {
	case StatusPending:
		return to == StatusPaid || to == StatusCanceled
	case StatusPaid:
		return to == StatusShipped || to == StatusCanceled
	case StatusShipped:
		return to == StatusDelivered || to == StatusCanceled
	case StatusDelivered, StatusCanceled:
		return false // 배송 완료 또는 취소 상태에서는 다른 상태로 전환 불가
	default:
		return false
	}
}