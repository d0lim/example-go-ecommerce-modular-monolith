package application

import (
	"context"

	"example.com/myapp/payment/domain"
)

// PaymentRepository는 결제 관련 영속성 인터페이스를 정의합니다.
type PaymentRepository interface {
	Save(ctx context.Context, payment *domain.Payment) error
	FindByID(ctx context.Context, id string) (*domain.Payment, error)
	FindByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
}

// PaymentGateway는 외부 결제 게이트웨이와의 통합을 정의합니다.
type PaymentGateway interface {
	ProcessPayment(ctx context.Context, payment *domain.Payment) (string, error)
	RefundPayment(ctx context.Context, payment *domain.Payment, reason string) error
}

// PaymentService는 결제 관련 비즈니스 로직을 정의합니다.
type PaymentService interface {
	CreatePayment(ctx context.Context, orderID string, amount float64, method domain.PaymentMethod, paymentData map[string]string) (*domain.Payment, error)
	ProcessPayment(ctx context.Context, paymentID string) (*domain.Payment, error)
	GetPayment(ctx context.Context, id string) (*domain.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	RefundPayment(ctx context.Context, id string, reason string) (*domain.Payment, error)
}

// PaymentUseCase는 PaymentService 구현체를 정의합니다.
type PaymentUseCase struct {
	repo    PaymentRepository
	gateway PaymentGateway
}

// NewPaymentUseCase는 새로운 PaymentUseCase 인스턴스를 생성합니다.
func NewPaymentUseCase(repo PaymentRepository, gateway PaymentGateway) *PaymentUseCase {
	return &PaymentUseCase{
		repo:    repo,
		gateway: gateway,
	}
}