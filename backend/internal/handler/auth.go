package handler

import (
	"net/http"

	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Verificar maneja POST /api/auth/verificar
func (h *AuthHandler) Verificar(c *gin.Context) {
	var req service.VerificarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(service.NewErrValidacion("El formato de la solicitud no es válido"))
		return
	}

	res, err := h.svc.Verificar(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Registrar maneja POST /api/auth/registro
func (h *AuthHandler) Registrar(c *gin.Context) {
	var req service.RegistrarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(service.NewErrValidacion("El formato de la solicitud no es válido"))
		return
	}

	if err := h.svc.Registrar(c.Request.Context(), req); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"mensaje": "Postulante registrado exitosamente"})
}

// Login maneja POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(service.NewErrValidacion("El formato de la solicitud no es válido"))
		return
	}

	res, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Recuperar maneja POST /api/auth/recuperar
func (h *AuthHandler) Recuperar(c *gin.Context) {
	var req service.RecuperarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(service.NewErrValidacion("El formato de la solicitud no es válido"))
		return
	}

	if err := h.svc.Recuperar(c.Request.Context(), req); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensaje": "Si los datos ingresados coinciden con un postulante registrado, se ha enviado un correo con instrucciones para restablecer su contraseña.",
	})
}

// Restablecer maneja POST /api/auth/restablecer
func (h *AuthHandler) Restablecer(c *gin.Context) {
	var req service.RestablecerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(service.NewErrValidacion("El formato de la solicitud no es válido"))
		return
	}

	if err := h.svc.Restablecer(c.Request.Context(), req); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensaje": "Contraseña restablecida exitosamente"})
}
