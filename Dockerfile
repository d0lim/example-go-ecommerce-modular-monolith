############################
# 빌드 스테이지
############################
FROM golang:1.21-alpine AS build

# 작업 디렉토리 설정
WORKDIR /app

# 필요한 도구 설치
RUN apk add --no-cache git

# 모듈 다운로드를 위한 Go 관련 파일 복사 
COPY go.work go.work
COPY shared/go.mod shared/go.mod
COPY internal/member/go.mod internal/member/go.mod
COPY internal/order/go.mod internal/order/go.mod
COPY internal/payment/go.mod internal/payment/go.mod

# 소스 코드 복사
COPY shared/ shared/
COPY internal/ internal/
COPY cmd/ cmd/

# 빌드
RUN go work sync
RUN CGO_ENABLED=0 GOOS=linux go build -o service ./cmd/service

############################
# 최종 이미지 생성
############################
FROM alpine:latest

# 보안 관련
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 필요한 패키지 설치
RUN apk --no-cache add ca-certificates tzdata

# 작업 디렉토리 설정
WORKDIR /app

# 빌드 스테이지에서 필요한 파일만 복사
COPY --from=build /app/service .
COPY configs/config.yaml configs/config.yaml
COPY api/ api/

# 소유권 변경
RUN chown -R appuser:appgroup /app

# 비특권 사용자로 전환
USER appuser

# 환경 변수 설정
ENV PORT=8080 \
    ENVIRONMENT=production \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=myapp \
    DB_SSLMODE=disable

# 애플리케이션 실행
CMD ["./service"]

# 포트 노출
EXPOSE 8080