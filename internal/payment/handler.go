package payment

import (
	"errors"
	"net/http"

	"flipay/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler is responsible only for HTTP concerns: binding, status codes, and JSON responses.
type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// Create godoc
// @Summary Create payment
// @Description Create a payment, generate VA/QRIS data, persist idempotency, and enqueue async processing.
// @Tags Payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string false "Unique key for safe request retry"
// @Param request body CreatePaymentRequest true "Payment payload"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Router /api/v1/payments [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	res, err := h.service.Create(c.Request.Context(), c.GetString("user_id"), c.GetHeader("Idempotency-Key"), req)
	if err != nil {
		if errors.Is(err, ErrIdempotencyConflict) {
			utils.Error(c, http.StatusConflict, err.Error())
			return
		}
		h.logger.Error("create payment failed", zap.Error(err))
		utils.Error(c, http.StatusInternalServerError, "failed to create payment")
		return
	}

	statusCode := http.StatusCreated
	message := "payment created"
	if res.Idempotent {
		statusCode = http.StatusOK
		message = "payment replayed from idempotency key"
	}
	utils.Success(c, statusCode, message, res)
}

// Detail godoc
// @Summary Get payment detail
// @Description Get a single payment owned by the authenticated user.
// @Tags Payments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /api/v1/payments/{id} [get]
func (h *Handler) Detail(c *gin.Context) {
	res, err := h.service.FindByID(c.Request.Context(), c.Param("id"), c.GetString("user_id"))
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			utils.Error(c, http.StatusNotFound, err.Error())
			return
		}
		h.logger.Error("get payment detail failed", zap.Error(err))
		utils.Error(c, http.StatusInternalServerError, "failed to get payment detail")
		return
	}
	utils.Success(c, http.StatusOK, "payment detail", res)
}

// History godoc
// @Summary Get payment history
// @Description List payments owned by the authenticated user.
// @Tags Payments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /api/v1/payments/history [get]
func (h *Handler) History(c *gin.Context) {
	res, err := h.service.History(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		h.logger.Error("get payment history failed", zap.Error(err))
		utils.Error(c, http.StatusInternalServerError, "failed to get payment history")
		return
	}
	utils.Success(c, http.StatusOK, "payment history", res)
}
