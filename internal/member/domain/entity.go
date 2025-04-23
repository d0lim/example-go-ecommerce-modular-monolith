package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidName     = errors.New("invalid name")
	ErrInvalidPassword = errors.New("invalid password")
)

// Member는 회원 엔티티를 나타냅니다.
// 캡슐화를 위해 모든 필드는 소문자(비공개)로 정의되어 있습니다.
type Member struct {
	id        string
	email     string
	name      string
	password  string // 실제로는 해시된 비밀번호가 저장됩니다
	createdAt time.Time
	updatedAt time.Time
}

// NewMember는 새로운 회원을 생성합니다.
func NewMember(email, name, password string) (*Member, error) {
	if email == "" {
		return nil, ErrInvalidEmail
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	now := time.Now()
	return &Member{
		id:        uuid.New().String(),
		email:     email,
		name:      name,
		password:  password, // 실제로는 해시 처리가 필요합니다
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ID는 회원의 고유 식별자를 반환합니다.
func (m *Member) ID() string {
	return m.id
}

// Email은 회원의 이메일 주소를 반환합니다.
func (m *Member) Email() string {
	return m.email
}

// Name은 회원의 이름을 반환합니다.
func (m *Member) Name() string {
	return m.name
}

// UpdateName은 회원의 이름을 업데이트합니다.
func (m *Member) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	m.name = name
	m.updatedAt = time.Now()
	return nil
}

// CreatedAt은 회원이 생성된 시간을 반환합니다.
func (m *Member) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt은 회원 정보가 마지막으로 업데이트된 시간을 반환합니다.
func (m *Member) UpdatedAt() time.Time {
	return m.updatedAt
}

// VerifyPassword는 제공된 비밀번호가 회원의 비밀번호와 일치하는지 확인합니다.
func (m *Member) VerifyPassword(password string) bool {
	// 실제로는 해시 비교 로직이 필요합니다
	return m.password == password
}