package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PaymentStatus는 결제 상태를 정의합니다.
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusApproved PaymentStatus = "approved"
	PaymentStatusRejected PaymentStatus = "rejected"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

// PaymentMethod는 결제 방법을 정의합니다.
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodVirtualAccount PaymentMethod = "virtual_account"
)

var (
	ErrInvalidPaymentAmount = errors.New("invalid payment amount")
	ErrInvalidOrderID       = errors.New("invalid order ID")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrPaymentNotFound      = errors.New("payment not found")
)

// Payment는 결제 엔티티를 나타냅니다.
type Payment struct {
	id            string
	orderID       string
	amount        float64
	method        PaymentMethod
	status        PaymentStatus
	transactionID string
	paymentData   map[string]string // 결제 방법별 추가 데이터
	createdAt     time.Time
	updatedAt     time.Time
}

// NewPayment는 새로운 결제를 생성합니다.
func NewPayment(orderID string, amount float64, method PaymentMethod, paymentData map[string]string) (*Payment, error) {
	if orderID == "" {
		return nil, ErrInvalidOrderID
	}
	if amount <= 0 {
		return nil, ErrInvalidPaymentAmount
	}
	if method == "" {
		return nil, ErrInvalidPaymentMethod
	}

	now := time.Now()
	return &Payment{
		id:          uuid.New().String(),
		orderID:     orderID,
		amount:      amount,
		method:      method,
		status:      PaymentStatusPending,
		paymentData: paymentData,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ID는 결제의 고유 식별자를 반환합니다.
func (p *Payment) ID() string {
	return p.id
}

// OrderID는 주문 ID를 반환합니다.
func (p *Payment) OrderID() string {
	return p.orderID
}

// Amount는 결제 금액을 반환합니다.
func (p *Payment) Amount() float64 {
	return p.amount
}

// Method는 결제 방법을 반환합니다.
func (p *Payment) Method() PaymentMethod {
	return p.method
}

// Status는 결제 상태를 반환합니다.
func (p *Payment) Status() PaymentStatus {
	return p.status
}

// TransactionID는 외부 결제 시스템의 트랜잭션 ID를 반환합니다.
func (p *Payment) TransactionID() string {
	return p.transactionID
}

// PaymentData는 결제 관련 추가 데이터를 반환합니다.
func (p *Payment) PaymentData() map[string]string {
	return p.paymentData
}

// CreatedAt은 결제가 생성된 시간을 반환합니다.
func (p *Payment) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt은 결제 정보가 마지막으로 업데이트된 시간을 반환합니다.
func (p *Payment) UpdatedAt() time.Time {
	return p.updatedAt
}

// SetTransactionID는 외부 결제 시스템의 트랜잭션 ID를 설정합니다.
func (p *Payment) SetTransactionID(transactionID string) {
	p.transactionID = transactionID
	p.updatedAt = time.Now()
}

// Approve는 결제를 승인 상태로 변경합니다.
func (p *Payment) Approve(transactionID string) {
	p.status = PaymentStatusApproved
	p.transactionID = transactionID
	p.updatedAt = time.Now()
}

// Reject는 결제를 거부 상태로 변경합니다.
func (p *Payment) Reject(reason string) {
	p.status = PaymentStatusRejected
	p.paymentData["reject_reason"] = reason
	p.updatedAt = time.Now()
}

// Refund는 결제를 환불 상태로 변경합니다.
func (p *Payment) Refund(reason string) error {
	if p.status != PaymentStatusApproved {
		return errors.New("only approved payments can be refunded")
	}
	
	p.status = PaymentStatusRefunded
	p.paymentData["refund_reason"] = reason
	p.updatedAt = time.Now()
	return nil
}