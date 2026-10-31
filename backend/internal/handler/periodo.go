package handler

import (
	"log"
	"net/http"

	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// PeriodoHandler maneja las peticiones HTTP del periodo de postulación.
type PeriodoHandler struct {
	svc *service.PeriodoService
}

// NewPeriodoHandler crea una instancia de PeriodoHandler.
func NewPeriodoHandler(svc *service.PeriodoService) *PeriodoHandler {
	return &PeriodoHandler{svc: svc}
}

// ObtenerPeriodo maneja GET /api/periodo — devuelve {abierto, inicio, fin} (T2).
func (h *PeriodoHandler) ObtenerPeriodo(c *gin.Context) {
	periodo, err := h.svc.ObtenerPeriodo(c.Request.Context())
	if err != nil {
		log.Printf("[ERROR] al obtener periodo de postulación: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al consultar el periodo."})
		return
	}

	c.JSON(http.StatusOK, periodo)
}
