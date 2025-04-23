package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"example.com/myapp/member/application"
	"example.com/myapp/member/domain"
	"example.com/myapp/shared/db"
	"github.com/jackc/pgx/v4"
)

// ErrMemberNotFound는 회원을 찾을 수 없을 때 발생하는 오류입니다.
var ErrMemberNotFound = errors.New("member not found")

// PostgresMemberRepository는 PostgreSQL을 사용하는 회원 저장소 구현체입니다.
type PostgresMemberRepository struct {
	db *db.Database
}

// NewPostgresMemberRepository는 새로운 PostgresMemberRepository 인스턴스를 생성합니다.
func NewPostgresMemberRepository(database *db.Database) application.MemberRepository {
	return &PostgresMemberRepository{
		db: database,
	}
}

// Save는 회원 정보를 데이터베이스에 저장합니다.
func (r *PostgresMemberRepository) Save(ctx context.Context, member *domain.Member) error {
	query := `
		INSERT INTO members (id, email, name, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		member.ID(),
		member.Email(),
		member.Name(),
		member.Password(),
		member.CreatedAt(),
		member.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save member: %w", err)
	}

	return nil
}

// FindByID는 ID로 회원을 조회합니다.
func (r *PostgresMemberRepository) FindByID(ctx context.Context, id string) (*domain.Member, error) {
	query := `
		SELECT id, email, name, password, created_at, updated_at
		FROM members
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)

	var memberID, email, name, password string
	var createdAt, updatedAt string

	err := row.Scan(&memberID, &email, &name, &password, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("failed to find member by ID: %w", err)
	}

	// 실제 구현에서는 DB 레코드를 도메인 엔티티로 변환하는 로직이 필요합니다.
	// 여기서는 코드 예시를 간략하게 하기 위해 생략합니다.
	return &domain.Member{}, nil
}

// FindByEmail은 이메일로 회원을 조회합니다.
func (r *PostgresMemberRepository) FindByEmail(ctx context.Context, email string) (*domain.Member, error) {
	query := `
		SELECT id, email, name, password, created_at, updated_at
		FROM members
		WHERE email = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, email)

	var memberID, memberEmail, name, password string
	var createdAt, updatedAt string

	err := row.Scan(&memberID, &memberEmail, &name, &password, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("failed to find member by email: %w", err)
	}

	// 실제 구현에서는 DB 레코드를 도메인 엔티티로 변환하는 로직이 필요합니다.
	return &domain.Member{}, nil
}

// Update는 회원 정보를 업데이트합니다.
func (r *PostgresMemberRepository) Update(ctx context.Context, member *domain.Member) error {
	query := `
		UPDATE members
		SET name = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		member.Name(),
		member.UpdatedAt(),
		member.ID(),
	)

	if err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}

	return nil
}

// Delete는 회원을 삭제합니다.
func (r *PostgresMemberRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM members
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}

	return nil
}