//go:build integration
// +build integration

package member

import (
	"context"
	"os"
	"testing"

	"example.com/myapp/member/application"
	"example.com/myapp/member/infrastructure"
	"example.com/myapp/shared/db"
)

func setupTestDatabase(t *testing.T) *db.Database {
	// 환경 변수에서 테스트 DB 정보 가져오기
	config := db.Config{
		Host:     os.Getenv("TEST_DB_HOST"),
		Port:     os.Getenv("TEST_DB_PORT"),
		User:     os.Getenv("TEST_DB_USER"),
		Password: os.Getenv("TEST_DB_PASSWORD"),
		DBName:   os.Getenv("TEST_DB_NAME"),
		SSLMode:  "disable",
	}

	// 기본값 설정
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == "" {
		config.Port = "5432"
	}
	if config.User == "" {
		config.User = "postgres"
	}
	if config.Password == "" {
		config.Password = "postgres"
	}
	if config.DBName == "" {
		config.DBName = "myapp_test"
	}

	database, err := db.NewDatabase(config)
	if err != nil {
		t.Fatalf("테스트 데이터베이스 연결 실패: %v", err)
	}

	// 테스트 테이블 초기화
	_, err = database.Pool.Exec(context.Background(), `
		TRUNCATE TABLE members CASCADE;
	`)
	if err != nil {
		t.Fatalf("테이블 초기화 실패: %v", err)
	}

	return database
}

func TestMemberIntegration(t *testing.T) {
	database := setupTestDatabase(t)
	defer database.Close()

	// 실제 저장소 및 유스케이스 생성
	repo := infrastructure.NewPostgresMemberRepository(database)
	useCase := application.NewMemberUseCase(repo)

	// 테스트 회원 정보
	email := "integration-test@example.com"
	name := "통합테스트"
	password := "password123"

	// 1. 회원 생성 테스트
	t.Run("회원 생성", func(t *testing.T) {
		member, err := useCase.CreateMember(context.Background(), email, name, password)
		if err != nil {
			t.Fatalf("회원 생성 실패: %v", err)
		}

		if member.Email() != email {
			t.Errorf("잘못된 이메일: got %v, want %v", member.Email(), email)
		}
		if member.Name() != name {
			t.Errorf("잘못된 이름: got %v, want %v", member.Name(), name)
		}
		
		// 생성된 ID 확인
		if member.ID() == "" {
			t.Error("ID가 생성되지 않았습니다")
		}
	})

	// 2. 회원 조회 테스트
	t.Run("회원 이메일로 조회", func(t *testing.T) {
		member, err := repo.FindByEmail(context.Background(), email)
		if err != nil {
			t.Fatalf("회원 조회 실패: %v", err)
		}

		if member.Email() != email {
			t.Errorf("잘못된 이메일: got %v, want %v", member.Email(), email)
		}
		if member.Name() != name {
			t.Errorf("잘못된 이름: got %v, want %v", member.Name(), name)
		}
	})

	// 3. 회원 업데이트 테스트
	t.Run("회원 이름 업데이트", func(t *testing.T) {
		// 먼저 회원 ID 조회
		existingMember, err := repo.FindByEmail(context.Background(), email)
		if err != nil {
			t.Fatalf("회원 조회 실패: %v", err)
		}

		// 이름 업데이트
		newName := "업데이트된이름"
		updatedMember, err := useCase.UpdateMember(context.Background(), existingMember.ID(), newName)
		if err != nil {
			t.Fatalf("회원 업데이트 실패: %v", err)
		}

		if updatedMember.Name() != newName {
			t.Errorf("이름 업데이트 실패: got %v, want %v", updatedMember.Name(), newName)
		}

		// DB에서 다시 조회하여 확인
		refetchedMember, err := repo.FindByID(context.Background(), existingMember.ID())
		if err != nil {
			t.Fatalf("업데이트 후 회원 조회 실패: %v", err)
		}

		if refetchedMember.Name() != newName {
			t.Errorf("DB에 이름 업데이트가 반영되지 않음: got %v, want %v", refetchedMember.Name(), newName)
		}
	})

	// 4. 회원 삭제 테스트
	t.Run("회원 삭제", func(t *testing.T) {
		// 먼저 회원 ID 조회
		existingMember, err := repo.FindByEmail(context.Background(), email)
		if err != nil {
			t.Fatalf("회원 조회 실패: %v", err)
		}

		// 회원 삭제
		err = useCase.DeleteMember(context.Background(), existingMember.ID())
		if err != nil {
			t.Fatalf("회원 삭제 실패: %v", err)
		}

		// 삭제 확인
		_, err = repo.FindByID(context.Background(), existingMember.ID())
		if err == nil {
			t.Error("회원이 삭제되지 않았습니다")
		}
	})
}