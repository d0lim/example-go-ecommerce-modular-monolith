package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/myapp/member/application"
	memberInfra "example.com/myapp/member/infrastructure"
	"example.com/myapp/order/application"
	orderInfra "example.com/myapp/order/infrastructure"
	"example.com/myapp/payment/application"
	paymentInfra "example.com/myapp/payment/infrastructure"
	"example.com/myapp/shared/db"
	"example.com/myapp/shared/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// 결제 게이트웨이 모의 구현
type DummyPaymentGateway struct{}

func (g *DummyPaymentGateway) ProcessPayment(ctx context.Context, payment *payment.domain.Payment) (string, error) {
	// 실제 구현에서는 외부 결제 API를 호출합니다
	return fmt.Sprintf("txn_%s", payment.ID()), nil
}

func (g *DummyPaymentGateway) RefundPayment(ctx context.Context, payment *payment.domain.Payment, reason string) error {
	// 실제 구현에서는 외부 결제 API를 호출합니다
	return nil
}

func main() {
	// 로거 초기화
	logger := log.NewLoggerFromEnv()
	logger.Info("서비스 시작 중...")

	// 데이터베이스 연결
	database, err := db.NewDatabaseFromEnv()
	if err != nil {
		logger.Fatalw("데이터베이스 연결 실패", "error", err)
	}
	defer database.Close()
	logger.Info("데이터베이스 연결 성공")

	// 저장소 초기화
	memberRepo := memberInfra.NewPostgresMemberRepository(database)
	orderRepo := orderInfra.NewPostgresOrderRepository(database)
	paymentRepo := paymentInfra.NewPostgresPaymentRepository(database)
	paymentGateway := &DummyPaymentGateway{}

	// 비즈니스 로직 유스케이스 초기화
	memberUseCase := member.NewMemberUseCase(memberRepo)
	orderUseCase := order.NewOrderUseCase(orderRepo)
	paymentUseCase := payment.NewPaymentUseCase(paymentRepo, paymentGateway)

	// Echo 인스턴스 생성
	e := echo.New()

	// 미들웨어 설정
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 요청 ID 미들웨어
	e.Use(middleware.RequestID())

	// API 라우팅 설정
	setupAPIRoutes(e, memberUseCase, orderUseCase, paymentUseCase, logger)

	// HTTP 서버 시작
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 서버 종료 처리를 위한 채널 설정
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 서버 시작
	go func() {
		address := fmt.Sprintf(":%s", port)
		logger.Infow("서버 시작", "address", address)
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			logger.Fatalw("서버 시작 실패", "error", err)
		}
	}()

	// 종료 신호 대기
	<-quit
	logger.Info("서버 종료 중...")

	// Graceful 종료
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Fatalw("서버 종료 실패", "error", err)
	}

	logger.Info("서버 종료 완료")
}

// setupAPIRoutes는 API 엔드포인트를 설정합니다.
func setupAPIRoutes(
	e *echo.Echo,
	memberUseCase member.MemberService,
	orderUseCase order.OrderService,
	paymentUseCase payment.PaymentService,
	logger *log.Logger,
) {
	// API 버전 그룹
	api := e.Group("/api/v1")

	// Health Check 엔드포인트
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// 회원 관련 엔드포인트
	members := api.Group("/members")
	members.POST("", createMemberHandler(memberUseCase, logger))
	members.GET("/:id", getMemberHandler(memberUseCase, logger))
	members.PUT("/:id", updateMemberHandler(memberUseCase, logger))
	members.DELETE("/:id", deleteMemberHandler(memberUseCase, logger))

	// 주문 관련 엔드포인트
	orders := api.Group("/orders")
	orders.POST("", createOrderHandler(orderUseCase, logger))
	orders.GET("/:id", getOrderHandler(orderUseCase, logger))
	orders.GET("/customer/:customerId", getCustomerOrdersHandler(orderUseCase, logger))
	orders.PUT("/:id/status", updateOrderStatusHandler(orderUseCase, logger))
	orders.POST("/:id/cancel", cancelOrderHandler(orderUseCase, logger))

	// 결제 관련 엔드포인트
	payments := api.Group("/payments")
	payments.POST("", createPaymentHandler(paymentUseCase, logger))
	payments.POST("/:id/process", processPaymentHandler(paymentUseCase, logger))
	payments.GET("/:id", getPaymentHandler(paymentUseCase, logger))
	payments.GET("/order/:orderId", getPaymentByOrderHandler(paymentUseCase, logger))
	payments.POST("/:id/refund", refundPaymentHandler(paymentUseCase, logger))
}

// API 핸들러 함수들 - 회원
func createMemberHandler(uc member.MemberService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		type request struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
		}

		var req request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		member, err := uc.CreateMember(c.Request().Context(), req.Email, req.Name, req.Password)
		if err != nil {
			logger.Errorw("회원 생성 실패", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"id":    member.ID(),
			"email": member.Email(),
			"name":  member.Name(),
		})
	}
}

func getMemberHandler(uc member.MemberService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID"})
		}

		member, err := uc.GetMember(c.Request().Context(), id)
		if err != nil {
			logger.Errorw("회원 조회 실패", "error", err, "id", id)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Member not found"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":    member.ID(),
			"email": member.Email(),
			"name":  member.Name(),
		})
	}
}

func updateMemberHandler(uc member.MemberService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID"})
		}

		type request struct {
			Name string `json:"name"`
		}

		var req request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		member, err := uc.UpdateMember(c.Request().Context(), id, req.Name)
		if err != nil {
			logger.Errorw("회원 업데이트 실패", "error", err, "id", id)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":    member.ID(),
			"email": member.Email(),
			"name":  member.Name(),
		})
	}
}

func deleteMemberHandler(uc member.MemberService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID"})
		}

		err := uc.DeleteMember(c.Request().Context(), id)
		if err != nil {
			logger.Errorw("회원 삭제 실패", "error", err, "id", id)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// API 핸들러 함수들 - 주문
func createOrderHandler(uc order.OrderService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		type orderItemRequest struct {
			ProductID string  `json:"productId"`
			Name      string  `json:"name"`
			Price     float64 `json:"price"`
			Quantity  int     `json:"quantity"`
		}

		type request struct {
			CustomerID string             `json:"customerId"`
			Items      []orderItemRequest `json:"items"`
		}

		var req request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		// 요청 데이터 변환
		items := make([]order.OrderItemRequest, len(req.Items))
		for i, item := range req.Items {
			items[i] = order.OrderItemRequest{
				ProductID: item.ProductID,
				Name:      item.Name,
				Price:     item.Price,
				Quantity:  item.Quantity,
			}
		}

		// 주문 생성
		newOrder, err := uc.CreateOrder(c.Request().Context(), req.CustomerID, items)
		if err != nil {
			logger.Errorw("주문 생성 실패", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"id":        newOrder.ID(),
			"customerId": newOrder.CustomerID(),
			"status":    string(newOrder.Status()),
			"total":     newOrder.TotalAmount(),
		})
	}
}

func getOrderHandler(uc order.OrderService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID"})
		}

		order, err := uc.GetOrder(c.Request().Context(), id)
		if err != nil {
			logger.Errorw("주문 조회 실패", "error", err, "id", id)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Order not found"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":        order.ID(),
			"customerId": order.CustomerID(),
			"status":    string(order.Status()),
			"total":     order.TotalAmount(),
		})
	}
}

func getCustomerOrdersHandler(uc order.OrderService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		customerID := c.Param("customerId")
		if customerID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing customer ID"})
		}

		orders, err := uc.GetCustomerOrders(c.Request().Context(), customerID)
		if err != nil {
			logger.Errorw("고객 주문 조회 실패", "error", err, "customerId", customerID)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// 응답 변환
		response := make([]map[string]interface{}, len(orders))
		for i, order := range orders {
			response[i] = map[string]interface{}{
				"id":        order.ID(),
				"customerId": order.CustomerID(),
				"status":    string(order.Status()),
				"total":     order.TotalAmount(),
			}
		}

		return c.JSON(http.StatusOK, response)
	}
}

func updateOrderStatusHandler(uc order.OrderService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID"})
		}

		type request struct {
			Status string `json:"status"`
		}

		var req request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		// 상태 변환
		status := order.domain.OrderStatus(req.Status)

		// 주문 상태 업데이트
		updatedOrder, err := uc.UpdateOrderStatus(c.Request().Context(), id, status)
		if err != nil {
			logger.Errorw("주문 상태 업데이트 실패", "error", err, "id", id, "status", status)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":        updatedOrder.ID(),
			"customerId": updatedOrder.CustomerID(),
			"status":    string(updatedOrder.Status()),
			"total":     updatedOrder.TotalAmount(),
		})
	}
}

func cancelOrderHandler(uc order.OrderService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID"})
		}

		// 주문 취소
		canceledOrder, err := uc.CancelOrder(c.Request().Context(), id)
		if err != nil {
			logger.Errorw("주문 취소 실패", "error", err, "id", id)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":        canceledOrder.ID(),
			"customerId": canceledOrder.CustomerID(),
			"status":    string(canceledOrder.Status()),
			"total":     canceledOrder.TotalAmount(),
		})
	}
}

// API 핸들러 함수들 - 결제
func createPaymentHandler(uc payment.PaymentService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		type request struct {
			OrderID     string                 `json:"orderId"`
			Amount      float64                `json:"amount"`
			Method      string                 `json:"method"`
			PaymentData map[string]string      `json:"paymentData"`
		}

		var req request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		// 결제 생성
		newPayment, err := uc.CreatePayment(
			c.Request().Context(),
			req.OrderID,
			req.Amount,
			payment.domain.PaymentMethod(req.Method),
			req.PaymentData,
		)
		if err != nil {
			logger.Errorw("결제 생성 실패", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"id":       newPayment.ID(),
			"orderId":  newPayment.OrderID(),
			"amount":   newPayment.Amount(),
			"method":   string(newPayment.Method()),
			"status":   string(newPayment.Status()),
		})
	}
}

func processPaymentHandler(uc payment.PaymentService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID"})
		}

		// 결제 처리
		processedPayment, err := uc.ProcessPayment(c.Request().Context(), id)
		if err != nil {
			logger.Errorw("결제 처리 실패", "error", err, "id", id)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":            processedPayment.ID(),
			"orderId":       processedPayment.OrderID(),
			"amount":        processedPayment.Amount(),
			"method":        string(processedPayment.Method()),
			"status":        string(processedPayment.Status()),
			"transactionId": processedPayment.TransactionID(),
		})
	}
}

func getPaymentHandler(uc payment.PaymentService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID"})
		}

		payment, err := uc.GetPayment(c.Request().Context(), id)
		if err != nil {
			logger.Errorw("결제 조회 실패", "error", err, "id", id)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Payment not found"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":            payment.ID(),
			"orderId":       payment.OrderID(),
			"amount":        payment.Amount(),
			"method":        string(payment.Method()),
			"status":        string(payment.Status()),
			"transactionId": payment.TransactionID(),
		})
	}
}

func getPaymentByOrderHandler(uc payment.PaymentService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		orderID := c.Param("orderId")
		if orderID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing order ID"})
		}

		payment, err := uc.GetPaymentByOrderID(c.Request().Context(), orderID)
		if err != nil {
			logger.Errorw("주문별 결제 조회 실패", "error", err, "orderId", orderID)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Payment not found"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":            payment.ID(),
			"orderId":       payment.OrderID(),
			"amount":        payment.Amount(),
			"method":        string(payment.Method()),
			"status":        string(payment.Status()),
			"transactionId": payment.TransactionID(),
		})
	}
}

func refundPaymentHandler(uc payment.PaymentService, logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID"})
		}

		type request struct {
			Reason string `json:"reason"`
		}

		var req request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		// 결제 환불 처리
		refundedPayment, err := uc.RefundPayment(c.Request().Context(), id, req.Reason)
		if err != nil {
			logger.Errorw("결제 환불 실패", "error", err, "id", id)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":      refundedPayment.ID(),
			"orderId": refundedPayment.OrderID(),
			"amount":  refundedPayment.Amount(),
			"status":  string(refundedPayment.Status()),
		})
	}
}