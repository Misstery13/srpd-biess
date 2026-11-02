package handler

import (
	"net/http"
	"strconv"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// PostulacionHandler maneja las peticiones de postulación (EF-12).
type PostulacionHandler struct {
	svc *service.PostulacionService
}

// NewPostulacionHandler crea una nueva instancia de PostulacionHandler.
func NewPostulacionHandler(svc *service.PostulacionService) *PostulacionHandler {
	return &PostulacionHandler{svc: svc}
}

func getPostulanteID(c *gin.Context) (int, error) {
	postulanteIDStr := c.GetString("postulante_id")
	return strconv.Atoi(postulanteIDStr)
}

// ValidarCargo maneja POST /api/postulacion/validar-cargo
func (h *PostulacionHandler) ValidarCargo(c *gin.Context) {
	postulanteID, err := getPostulanteID(c)
	if err != nil {
		_ = c.Error(service.NewErrNoAutorizado("Token de sesión inválido."))
		return
	}

	var req model.ValidarCargoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(service.NewErrValidacion("El formato de la solicitud es inválido. Debe enviar el campo 'cargo'."))
		return
	}

	resp, err := h.svc.ValidarCargo(c.Request.Context(), postulanteID, &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Obtener maneja GET /api/postulacion
func (h *PostulacionHandler) Obtener(c *gin.Context) {
	postulanteID, err := getPostulanteID(c)
	if err != nil {
		_ = c.Error(service.NewErrNoAutorizado("Token de sesión inválido."))
		return
	}

	postulacion, err := h.svc.ObtenerPostulacion(c.Request.Context(), postulanteID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, postulacion)
}

// Guardar maneja PUT /api/postulacion
func (h *PostulacionHandler) Guardar(c *gin.Context) {
	postulanteID, err := getPostulanteID(c)
	if err != nil {
		_ = c.Error(service.NewErrNoAutorizado("Token de sesión inválido."))
		return
	}

	var req model.PostulacionCompleta
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(service.NewErrValidacion("El cuerpo de la postulación contiene un formato JSON inválido."))
		return
	}

	codigo, err := h.svc.GuardarPostulacion(c.Request.Context(), postulanteID, &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"codigo": codigo})
}

// SubirDocumento maneja POST /api/postulacion/documentos
func (h *PostulacionHandler) SubirDocumento(c *gin.Context) {
	postulanteID, err := getPostulanteID(c)
	if err != nil {
		_ = c.Error(service.NewErrNoAutorizado("Token de sesión inválido."))
		return
	}

	tipo := c.PostForm("tipo")
	if tipo == "" {
		_ = c.Error(service.NewErrValidacion("Debe indicar el tipo de documento en el formulario."))
		return
	}

	fileHeader, err := c.FormFile("archivo")
	if err != nil {
		_ = c.Error(service.NewErrValidacion("Debe adjuntar un archivo en el campo 'archivo'."))
		return
	}

	ruta, err := h.svc.SubirDocumento(c.Request.Context(), postulanteID, tipo, fileHeader)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ruta": ruta})
}

// Enviar maneja POST /api/postulacion/enviar
func (h *PostulacionHandler) Enviar(c *gin.Context) {
	postulanteID, err := getPostulanteID(c)
	if err != nil {
		_ = c.Error(service.NewErrNoAutorizado("Token de sesión inválido."))
		return
	}

	var req model.EnviarPostulacionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(service.NewErrValidacion("Formato de solicitud inválido."))
		return
	}

	codigo, err := h.svc.EnviarPostulacion(c.Request.Context(), postulanteID, &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensaje": "Postulación enviada exitosamente",
		"codigo":  codigo,
	})
}
