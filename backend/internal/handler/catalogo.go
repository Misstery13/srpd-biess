package handler

import (
	"net/http"
	"strconv"

	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type CatalogoHandler struct {
	svc service.CatalogoService
}

func NewCatalogoHandler(svc service.CatalogoService) *CatalogoHandler {
	return &CatalogoHandler{svc: svc}
}

// ListarPaises maneja GET /api/catalogos/paises
func (h *CatalogoHandler) ListarPaises(c *gin.Context) {
	paises, err := h.svc.ObtenerPaises(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, paises)
}

// ListarProvincias maneja GET /api/catalogos/paises/:id/provincias
func (h *CatalogoHandler) ListarProvincias(c *gin.Context) {
	idStr := c.Param("id")
	paisID, err := strconv.Atoi(idStr)
	if err != nil || paisID <= 0 {
		_ = c.Error(service.NewErrValidacion("El ID de país debe ser un número entero positivo"))
		return
	}

	provincias, err := h.svc.ObtenerProvincias(c.Request.Context(), paisID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, provincias)
}

// ListarCantones maneja GET /api/catalogos/provincias/:id/cantones
func (h *CatalogoHandler) ListarCantones(c *gin.Context) {
	idStr := c.Param("id")
	provinciaID, err := strconv.Atoi(idStr)
	if err != nil || provinciaID <= 0 {
		_ = c.Error(service.NewErrValidacion("El ID de provincia debe ser un número entero positivo"))
		return
	}

	cantones, err := h.svc.ObtenerCantones(c.Request.Context(), provinciaID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, cantones)
}
