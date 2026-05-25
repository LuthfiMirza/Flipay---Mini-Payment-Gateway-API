package webhook

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	secret string
	logger *zap.Logger
}

func NewHandler(secret string, logger *zap.Logger) *Handler {
	return &Handler{secret: secret, logger: logger}
}

// PaymentCallback is a merchant callback receiver simulation.
func (h *Handler) PaymentCallback(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid payload"})
		return
	}
	if !ValidateSignature(payload, h.secret, c.GetHeader("X-Flipay-Signature")) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid signature"})
		return
	}
	h.logger.Info("callback received", zap.ByteString("payload", payload))
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "callback received"})
}
