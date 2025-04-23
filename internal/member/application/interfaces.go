package application

import (
	"context"

	"example.com/myapp/member/domain"
)

// MemberRepository는 회원 관련 영속성 인터페이스를 정의합니다.
type MemberRepository interface {
	Save(ctx context.Context, member *domain.Member) error
	FindByID(ctx context.Context, id string) (*domain.Member, error)
	FindByEmail(ctx context.Context, email string) (*domain.Member, error)
	Update(ctx context.Context, member *domain.Member) error
	Delete(ctx context.Context, id string) error
}

// MemberService는 회원 관련 비즈니스 로직을 정의합니다.
type MemberService interface {
	CreateMember(ctx context.Context, email, name, password string) (*domain.Member, error)
	GetMember(ctx context.Context, id string) (*domain.Member, error)
	UpdateMember(ctx context.Context, id, name string) (*domain.Member, error)
	DeleteMember(ctx context.Context, id string) error
}

// MemberUseCase는 MemberService 구현체를 정의합니다.
type MemberUseCase struct {
	repo MemberRepository
}

// NewMemberUseCase는 새로운 MemberUseCase 인스턴스를 생성합니다.
func NewMemberUseCase(repo MemberRepository) *MemberUseCase {
	return &MemberUseCase{
		repo: repo,
	}
}