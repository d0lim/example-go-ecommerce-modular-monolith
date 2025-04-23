package application

import (
	"context"
	"testing"

	"example.com/myapp/member/domain"
)

// FakeMemberRepository는 테스트를 위한 가짜 MemberRepository 구현체입니다.
type FakeMemberRepository struct {
	members map[string]*domain.Member
	emails  map[string]*domain.Member
}

// NewFakeMemberRepository는 새로운 FakeMemberRepository 인스턴스를 생성합니다.
func NewFakeMemberRepository() *FakeMemberRepository {
	return &FakeMemberRepository{
		members: make(map[string]*domain.Member),
		emails:  make(map[string]*domain.Member),
	}
}

func (f *FakeMemberRepository) Save(ctx context.Context, member *domain.Member) error {
	f.members[member.ID()] = member
	f.emails[member.Email()] = member
	return nil
}

func (f *FakeMemberRepository) FindByID(ctx context.Context, id string) (*domain.Member, error) {
	member, ok := f.members[id]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	return member, nil
}

func (f *FakeMemberRepository) FindByEmail(ctx context.Context, email string) (*domain.Member, error) {
	member, ok := f.emails[email]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	return member, nil
}

func (f *FakeMemberRepository) Update(ctx context.Context, member *domain.Member) error {
	f.members[member.ID()] = member
	return nil
}

func (f *FakeMemberRepository) Delete(ctx context.Context, id string) error {
	member, ok := f.members[id]
	if !ok {
		return domain.ErrMemberNotFound
	}
	
	delete(f.emails, member.Email())
	delete(f.members, id)
	return nil
}

func TestCreateMember(t *testing.T) {
	// 테스트 케이스
	tests := []struct {
		name     string
		email    string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "유효한 회원 생성",
			email:    "test@example.com",
			username: "테스트사용자",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "이메일 없음",
			email:    "",
			username: "테스트사용자",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "이름 없음",
			email:    "test@example.com",
			username: "",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "비밀번호 너무 짧음",
			email:    "test@example.com",
			username: "테스트사용자",
			password: "123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 가짜 저장소 준비
			repo := NewFakeMemberRepository()
			useCase := NewMemberUseCase(repo)

			// 테스트 실행
			member, err := useCase.CreateMember(context.Background(), tt.email, tt.username, tt.password)

			// 결과 확인
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if member == nil {
					t.Error("CreateMember() expected non-nil member")
					return
				}
				if member.Email() != tt.email {
					t.Errorf("CreateMember() email = %v, want %v", member.Email(), tt.email)
				}
				if member.Name() != tt.username {
					t.Errorf("CreateMember() name = %v, want %v", member.Name(), tt.username)
				}
			}
		})
	}
}

func TestCreateMemberWithDuplicateEmail(t *testing.T) {
	// 가짜 저장소 준비
	repo := NewFakeMemberRepository()
	useCase := NewMemberUseCase(repo)
	
	// 첫 번째 회원 생성
	email := "test@example.com"
	_, err := useCase.CreateMember(context.Background(), email, "테스트사용자1", "password123")
	if err != nil {
		t.Fatalf("첫 번째 사용자 생성 실패: %v", err)
	}
	
	// 동일한 이메일로 두 번째 회원 생성 시도
	_, err = useCase.CreateMember(context.Background(), email, "테스트사용자2", "password456")
	
	// 이메일 중복 에러 확인
	if err == nil {
		t.Error("중복 이메일 검사 실패: 에러가 발생해야 함")
	}
	if err != ErrMemberAlreadyExists {
		t.Errorf("잘못된 에러 타입: got %v, want %v", err, ErrMemberAlreadyExists)
	}
}