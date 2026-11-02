package handler

import (
	"net/http"
	"strconv"

	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// PostulanteHandler maneja los endpoints relacionados al postulante y su perfil (EF-12).
type PostulanteHandler struct {
	svc *service.PostulanteService
}

// NewPostulanteHandler crea una nueva instancia de PostulanteHandler.
func NewPostulanteHandler(svc *service.PostulanteService) *PostulanteHandler {
	return &PostulanteHandler{svc: svc}
}

// ObtenerPerfil maneja GET /api/postulante/me (V5, P6).
func (h *PostulanteHandler) ObtenerPerfil(c *gin.Context) {
	postulanteIDStr := c.GetString("postulante_id")
	postulanteID, err := strconv.Atoi(postulanteIDStr)
	if err != nil {
		_ = c.Error(service.NewErrNoAutorizado("Token de sesión inválido."))
		return
	}

	perfil, err := h.svc.ObtenerPerfil(c.Request.Context(), postulanteID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, perfil)
}

// SuprimirCuenta maneja DELETE /api/postulante/me (B4, L5).
func (h *PostulanteHandler) SuprimirCuenta(c *gin.Context) {
	postulanteIDStr := c.GetString("postulante_id")
	postulanteID, err := strconv.Atoi(postulanteIDStr)
	if err != nil {
		_ = c.Error(service.NewErrNoAutorizado("Token de sesión inválido."))
		return
	}

	if err := h.svc.SuprimirCuenta(c.Request.Context(), postulanteID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
