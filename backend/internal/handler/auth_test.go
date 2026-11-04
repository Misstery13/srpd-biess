package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type mockAuthService struct {
	verificarFunc   func(ctx context.Context, req service.VerificarRequest) (*service.VerificarResponse, error)
	registrarFunc   func(ctx context.Context, req service.RegistrarRequest) error
	loginFunc       func(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error)
	recuperarFunc   func(ctx context.Context, req service.RecuperarRequest) error
	restablecerFunc func(ctx context.Context, req service.RestablecerRequest) error
}

func (m *mockAuthService) Verificar(ctx context.Context, req service.VerificarRequest) (*service.VerificarResponse, error) {
	if m.verificarFunc != nil {
		return m.verificarFunc(ctx, req)
	}
	return &service.VerificarResponse{Nombres: "JUAN PÉREZ"}, nil
}

func (m *mockAuthService) Registrar(ctx context.Context, req service.RegistrarRequest) error {
	if m.registrarFunc != nil {
		return m.registrarFunc(ctx, req)
	}
	return nil
}

func (m *mockAuthService) Login(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return &service.LoginResponse{Token: "test.jwt.token"}, nil
}

func (m *mockAuthService) Recuperar(ctx context.Context, req service.RecuperarRequest) error {
	if m.recuperarFunc != nil {
		return m.recuperarFunc(ctx, req)
	}
	return nil
}

func (m *mockAuthService) Restablecer(ctx context.Context, req service.RestablecerRequest) error {
	if m.restablecerFunc != nil {
		return m.restablecerFunc(ctx, req)
	}
	return nil
}

type mockPermisoRepo struct {
	permisos map[string]bool
}

func (m *mockPermisoRepo) TienePermiso(ctx context.Context, rolCodigo, metodo, ruta string) (bool, error) {
	k := rolCodigo + ":" + metodo + ":" + ruta
	return m.permisos[k], nil
}

func setupAuthRouter(svc service.AuthService, permisoRepo *mockPermisoRepo, jwtSecret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeadersMiddleware())
	r.Use(ErrorMiddleware())

	authHdl := NewAuthHandler(svc)
	authGrp := r.Group("/api/auth")
	{
		authGrp.POST("/verificar", authHdl.Verificar)
		authGrp.POST("/registro", authHdl.Registrar)
		authGrp.POST("/login", authHdl.Login)
		authGrp.POST("/recuperar", authHdl.Recuperar)
		authGrp.POST("/restablecer", authHdl.Restablecer)
	}

	if permisoRepo != nil {
		protegido := r.Group("/api/postulante")
		protegido.Use(JWTMiddleware(jwtSecret))
		protegido.Use(RBACMiddleware(permisoRepo))
		{
			protegido.GET("/me", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
		}
	}

	return r
}

func TestAuthHandler_Verificar(t *testing.T) {
	mockSvc := &mockAuthService{}
	r := setupAuthRouter(mockSvc, nil, "12345678901234567890123456789012")

	body, _ := json.Marshal(map[string]string{
		"cedula":          "1710034065",
		"codigo_dactilar": "V4443V4444",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/verificar", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba status 200, obtuvo %d", w.Code)
	}
}

func TestAuthHandler_Registrar(t *testing.T) {
	mockSvc := &mockAuthService{}
	r := setupAuthRouter(mockSvc, nil, "12345678901234567890123456789012")

	body, _ := json.Marshal(map[string]string{
		"cedula":   "1710034065",
		"email":    "juan@example.com",
		"password": "Password123",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/registro", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("esperaba status 201, obtuvo %d", w.Code)
	}
}

func TestAuthHandler_Middlewares_JWT_Y_RBAC(t *testing.T) {
	jwtSecret := "12345678901234567890123456789012"
	permisoRepo := &mockPermisoRepo{
		permisos: map[string]bool{
			"POSTULANTE:GET:/api/postulante/me": true,
		},
	}
	r := setupAuthRouter(&mockAuthService{}, permisoRepo, jwtSecret)

	// 1. Petición sin token -> 401
	req1, _ := http.NewRequest(http.MethodGet, "/api/postulante/me", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusUnauthorized {
		t.Errorf("esperaba 401 sin token, obtuvo %d", w1.Code)
	}

	// 2. Token con rol POSTULANTE (tiene permiso) -> 200
	tokenValido := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "10",
		"rol": "POSTULANTE",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenStr, _ := tokenValido.SignedString([]byte(jwtSecret))

	req2, _ := http.NewRequest(http.MethodGet, "/api/postulante/me", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenStr)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("esperaba 200 con token y permiso, obtuvo %d", w2.Code)
	}

	// 3. Token con rol que NO tiene permiso -> 403 (R3)
	tokenSinPermiso := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "20",
		"rol": "OTRO_ROL",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenSinPermisoStr, _ := tokenSinPermiso.SignedString([]byte(jwtSecret))

	req3, _ := http.NewRequest(http.MethodGet, "/api/postulante/me", nil)
	req3.Header.Set("Authorization", "Bearer "+tokenSinPermisoStr)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusForbidden {
		t.Errorf("esperaba 403 sin permiso RBAC, obtuvo %d", w3.Code)
	}
}
