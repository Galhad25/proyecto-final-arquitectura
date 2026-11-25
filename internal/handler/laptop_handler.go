package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"biblioteca-api/internal/models"
	"biblioteca-api/internal/repository"
	"biblioteca-api/internal/service"
)

type LaptopHandler struct {
	laptopService *service.LaptopService
}

func NewLaptopHandler(laptopService *service.LaptopService) *LaptopHandler {
	return &LaptopHandler{laptopService: laptopService}
}

// List godoc
// @Summary      Listar laptops
// @Description  Devuelve el inventario de laptops, con filtro opcional por estado (available | loaned)
// @Tags         laptops
// @Produce      json
// @Security     BearerAuth
// @Param        status query string false "Filtrar por estado (available | loaned)"
// @Success      200 {array} models.Laptop
// @Router       /laptops [get]
func (h *LaptopHandler) List(c *gin.Context) {
	laptops, err := h.laptopService.List(c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo obtener el inventario de laptops"})
		return
	}
	c.JSON(http.StatusOK, laptops)
}

// Create godoc
// @Summary      Registrar una laptop (solo admin)
// @Tags         laptops
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.LaptopRequest true "Datos de la laptop"
// @Success      201 {object} models.Laptop
// @Failure      400 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Router       /laptops [post]
func (h *LaptopHandler) Create(c *gin.Context) {
	var req models.LaptopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	laptop, err := h.laptopService.Create(req)
	if errors.Is(err, repository.ErrDuplicateSerial) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo crear la laptop"})
		return
	}
	c.JSON(http.StatusCreated, laptop)
}
