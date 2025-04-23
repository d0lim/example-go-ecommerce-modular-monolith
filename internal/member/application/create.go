package application

import (
	"context"
	"errors"

	"example.com/myapp/member/domain"
)

var (
	ErrMemberAlreadyExists = errors.New("member already exists with this email")
)

// CreateMemberRequest는 회원 생성 요청 정보를 정의합니다.
type CreateMemberRequest struct {
	Email    string
	Name     string
	Password string
}

// CreateMemberHandler는 회원 생성 유스케이스를 구현합니다.
func (uc *MemberUseCase) CreateMember(ctx context.Context, email, name, password string) (*domain.Member, error) {
	// 1. 이메일 중복 검사
	existingMember, err := uc.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrMemberNotFound) {
		return nil, err
	}
	if existingMember != nil {
		return nil, ErrMemberAlreadyExists
	}

	// 2. 새 회원 생성
	member, err := domain.NewMember(email, name, password)
	if err != nil {
		return nil, err
	}

	// 3. 저장소에 회원 저장
	if err := uc.repo.Save(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// GetMember는 회원 정보를 조회합니다.
func (uc *MemberUseCase) GetMember(ctx context.Context, id string) (*domain.Member, error) {
	return uc.repo.FindByID(ctx, id)
}

// UpdateMember는 회원 정보를 업데이트합니다.
func (uc *MemberUseCase) UpdateMember(ctx context.Context, id, name string) (*domain.Member, error) {
	member, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := member.UpdateName(name); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// DeleteMember는 회원을 삭제합니다.
func (uc *MemberUseCase) DeleteMember(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}