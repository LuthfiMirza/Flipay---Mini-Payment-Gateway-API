package auth

import (
	"errors"
	"net/http"

	"flipay/internal/utils"
	"github.com/gin-gonic/gin"
)

// Handler translates HTTP requests into service calls.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	res, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			utils.Error(c, http.StatusConflict, err.Error())
			return
		}
		utils.Error(c, http.StatusInternalServerError, "failed to register user")
		return
	}
	utils.Success(c, http.StatusCreated, "user registered successfully", res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredential) {
			utils.Error(c, http.StatusUnauthorized, err.Error())
			return
		}
		utils.Error(c, http.StatusInternalServerError, "failed to login")
		return
	}
	utils.Success(c, http.StatusOK, "login successful", res)
}
