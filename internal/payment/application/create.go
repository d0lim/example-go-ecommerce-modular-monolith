package application

import (
	"context"
	"errors"
	"fmt"

	"example.com/myapp/payment/domain"
)

var (
	ErrPaymentAlreadyExists = errors.New("payment already exists for this order")
	ErrInvalidPaymentID     = errors.New("invalid payment ID")
)

// CreatePayment는 새로운 결제를 생성합니다.
func (uc *PaymentUseCase) CreatePayment(
	ctx context.Context,
	orderID string,
	amount float64,
	method domain.PaymentMethod,
	paymentData map[string]string,
) (*domain.Payment, error) {
	// 이미 해당 주문에 대한 결제가 있는지 확인
	existingPayment, err := uc.repo.FindByOrderID(ctx, orderID)
	if err != nil && !errors.Is(err, domain.ErrPaymentNotFound) {
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}
	if existingPayment != nil {
		return nil, ErrPaymentAlreadyExists
	}

	// 결제 엔티티 생성
	payment, err := domain.NewPayment(orderID, amount, method, paymentData)
	if err != nil {
		return nil, err
	}

	// 저장소에 결제 저장
	if err := uc.repo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return payment, nil
}

// ProcessPayment는 결제를 처리합니다.
func (uc *PaymentUseCase) ProcessPayment(ctx context.Context, paymentID string) (*domain.Payment, error) {
	if paymentID == "" {
		return nil, ErrInvalidPaymentID
	}

	// 결제 정보 조회
	payment, err := uc.repo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// 이미 처리된 결제인지 확인
	if payment.Status() != domain.PaymentStatusPending {
		return payment, nil
	}

	// 결제 게이트웨이를 통해 결제 처리
	transactionID, err := uc.gateway.ProcessPayment(ctx, payment)
	if err != nil {
		// 결제 실패 처리
		payment.Reject(err.Error())
		if updateErr := uc.repo.Update(ctx, payment); updateErr != nil {
			return nil, fmt.Errorf("failed to update payment status after rejection: %w", updateErr)
		}
		return payment, fmt.Errorf("payment processing failed: %w", err)
	}

	// 결제 성공 처리
	payment.Approve(transactionID)
	if err := uc.repo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment status after approval: %w", err)
	}

	return payment, nil
}

// GetPayment는 결제 ID로 결제 정보를 조회합니다.
func (uc *PaymentUseCase) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	if id == "" {
		return nil, ErrInvalidPaymentID
	}
	return uc.repo.FindByID(ctx, id)
}

// GetPaymentByOrderID는 주문 ID로 결제 정보를 조회합니다.
func (uc *PaymentUseCase) GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	if orderID == "" {
		return nil, domain.ErrInvalidOrderID
	}
	return uc.repo.FindByOrderID(ctx, orderID)
}

// RefundPayment는 결제를 환불합니다.
func (uc *PaymentUseCase) RefundPayment(ctx context.Context, id string, reason string) (*domain.Payment, error) {
	if id == "" {
		return nil, ErrInvalidPaymentID
	}

	// 결제 정보 조회
	payment, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 게이트웨이를 통해 환불 처리
	if err := uc.gateway.RefundPayment(ctx, payment, reason); err != nil {
		return nil, fmt.Errorf("refund processing failed: %w", err)
	}

	// 결제 상태 업데이트
	if err := payment.Refund(reason); err != nil {
		return nil, err
	}

	// 저장소 업데이트
	if err := uc.repo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment status after refund: %w", err)
	}

	return payment, nil
}