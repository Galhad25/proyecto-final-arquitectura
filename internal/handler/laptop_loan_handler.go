package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"biblioteca-api/internal/models"
	"biblioteca-api/internal/service"
)

type LaptopLoanHandler struct {
	loanService *service.LaptopLoanService
}

func NewLaptopLoanHandler(loanService *service.LaptopLoanService) *LaptopLoanHandler {
	return &LaptopLoanHandler{loanService: loanService}
}

// Borrow godoc
// @Summary      Pedir prestada una laptop
// @Tags         laptop-loans
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.LaptopLoanRequest true "ID de la laptop a pedir prestada"
// @Success      201 {object} models.LaptopLoan
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Router       /laptop-loans [post]
func (h *LaptopLoanHandler) Borrow(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req models.LaptopLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan, err := h.loanService.Borrow(userID, req)
	if err != nil {
		writeLaptopLoanError(c, err)
		return
	}
	c.JSON(http.StatusCreated, loan)
}

// MyLoans godoc
// @Summary      Mis préstamos de laptops
// @Description  Lista los préstamos de laptop del usuario autenticado (activos e histórico)
// @Tags         laptop-loans
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.LaptopLoanDetail
// @Router       /laptop-loans/me [get]
func (h *LaptopLoanHandler) MyLoans(c *gin.Context) {
	userID := c.GetInt("user_id")

	loans, err := h.loanService.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron obtener los préstamos de laptops"})
		return
	}
	c.JSON(http.StatusOK, loans)
}

// Return godoc
// @Summary      Devolver una laptop
// @Tags         laptop-loans
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID del préstamo de laptop"
// @Success      200 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Router       /laptop-loans/{id}/return [put]
func (h *LaptopLoanHandler) Return(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	userID := c.GetInt("user_id")
	role := c.GetString("role")

	if err := h.loanService.Return(id, userID, role); err != nil {
		writeLaptopLoanError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "laptop devuelta correctamente"})
}

// writeLaptopLoanError traduce los errores de negocio a estados HTTP.
func writeLaptopLoanError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrLaptopNotFound), errors.Is(err, service.ErrLaptopLoanNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrLaptopNotAvailable),
		errors.Is(err, service.ErrUserHasActiveLaptop),
		errors.Is(err, service.ErrLaptopLoanReturned):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrLaptopLoanNotYours):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno al procesar el préstamo de laptop"})
	}
}
