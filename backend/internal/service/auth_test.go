package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// Mock repos y cache para pruebas unitarias de AuthService
type mockRegistroCivilRepo struct {
	personas map[string]*model.RegistroCivil
}

func (m *mockRegistroCivilRepo) BuscarPorCedula(ctx context.Context, cedula string) (*model.RegistroCivil, error) {
	return m.personas[cedula], nil
}

type mockPostulanteRepo struct {
	postulantesPorCedula map[string]*model.Postulante
	postulantesPorEmail  map[string]*model.Postulante
}

func (m *mockPostulanteRepo) BuscarActivoPorCedula(ctx context.Context, cedula string) (*model.Postulante, error) {
	return m.postulantesPorCedula[cedula], nil
}

func (m *mockPostulanteRepo) BuscarActivoPorEmail(ctx context.Context, email string) (*model.Postulante, error) {
	return m.postulantesPorEmail[strings.ToLower(email)], nil
}

func (m *mockPostulanteRepo) Crear(ctx context.Context, p *model.Postulante) error {
	p.ID = 10
	p.CreatedAt = time.Now()
	m.postulantesPorCedula[p.Cedula] = p
	m.postulantesPorEmail[strings.ToLower(p.Email)] = p
	return nil
}

func (m *mockPostulanteRepo) ActualizarPassword(ctx context.Context, cedula, passwordHash string) error {
	p := m.postulantesPorCedula[cedula]
	if p == nil {
		return errors.New("postulante no encontrado")
	}
	p.PasswordHash = passwordHash
	return nil
}

func (m *mockPostulanteRepo) BuscarRolPorCodigo(ctx context.Context, codigo string) (*model.Rol, error) {
	return &model.Rol{ID: 1, Codigo: codigo, Nombre: "Postulante"}, nil
}

func (m *mockPostulanteRepo) ObtenerDetalleRegistroCivil(ctx context.Context, postulanteID int) (*model.PostulantePerfil, error) {
	return nil, nil
}

func (m *mockPostulanteRepo) SuprimirPostulante(ctx context.Context, postulanteID int) error {
	return nil
}

type mockAuthCache struct {
	intentos    map[string]int64
	bloqueados  map[string]bool
	verificados map[string]bool
	resetTokens map[string]string
}

func newMockAuthCache() *mockAuthCache {
	return &mockAuthCache{
		intentos:    make(map[string]int64),
		bloqueados:  make(map[string]bool),
		verificados: make(map[string]bool),
		resetTokens: make(map[string]string),
	}
}

func (m *mockAuthCache) IncrementarIntentos(ctx context.Context, tipo, cedula string) (int64, error) {
	k := tipo + ":" + cedula
	m.intentos[k]++
	if m.intentos[k] >= 3 {
		m.bloqueados[k] = true
	}
	return m.intentos[k], nil
}

func (m *mockAuthCache) EstaBloqueado(ctx context.Context, tipo, cedula string) (bool, error) {
	k := tipo + ":" + cedula
	return m.bloqueados[k], nil
}

func (m *mockAuthCache) ReiniciarIntentos(ctx context.Context, tipo, cedula string) error {
	k := tipo + ":" + cedula
	delete(m.intentos, k)
	delete(m.bloqueados, k)
	return nil
}

func (m *mockAuthCache) GuardarVerificado(ctx context.Context, cedula string) error {
	m.verificados[cedula] = true
	return nil
}

func (m *mockAuthCache) EsVerificado(ctx context.Context, cedula string) (bool, error) {
	return m.verificados[cedula], nil
}

func (m *mockAuthCache) BorrarVerificado(ctx context.Context, cedula string) error {
	delete(m.verificados, cedula)
	return nil
}

func (m *mockAuthCache) GuardarResetToken(ctx context.Context, token, cedula string) error {
	m.resetTokens[token] = cedula
	return nil
}

func (m *mockAuthCache) ObtenerYBorrarResetToken(ctx context.Context, token string) (string, error) {
	cedula := m.resetTokens[token]
	delete(m.resetTokens, token)
	return cedula, nil
}

type mockEmailService struct {
	enviados []string
}

func (m *mockEmailService) EnviarRecuperacion(destinatario, token string) error {
	m.enviados = append(m.enviados, destinatario)
	return nil
}

func (m *mockEmailService) EnviarConfirmacionPostulacion(destinatario, codigo string) error {
	m.enviados = append(m.enviados, destinatario)
	return nil
}

func setupAuthService() (AuthService, *mockAuthCache, *mockPostulanteRepo) {
	rcRepo := &mockRegistroCivilRepo{
		personas: map[string]*model.RegistroCivil{
			"1710034065": {
				Cedula:          "1710034065",
				Nombres:         "JUAN CARLOS",
				Apellidos:       "PÉREZ LÓPEZ",
				FechaExpedicion: time.Date(2020, 3, 15, 0, 0, 0, 0, time.UTC),
				CodigoDactilar:  "V4443V4444",
			},
		},
	}

	postRepo := &mockPostulanteRepo{
		postulantesPorCedula: make(map[string]*model.Postulante),
		postulantesPorEmail:  make(map[string]*model.Postulante),
	}

	cache := newMockAuthCache()
	emailSvc := &mockEmailService{}
	jwtSecret := "12345678901234567890123456789012"

	svc := NewAuthService(rcRepo, postRepo, cache, emailSvc, jwtSecret)
	return svc, cache, postRepo
}

func TestAuthService_Verificar(t *testing.T) {
	svc, cache, _ := setupAuthService()
	ctx := context.Background()

	dactilarValido := "V4443V4444"
	dactilarInvalido := "X9999X9999"

	// 1. Verificación exitosa con código dactilar
	res, err := svc.Verificar(ctx, VerificarRequest{
		Cedula:         "1710034065",
		CodigoDactilar: &dactilarValido,
	})
	if err != nil {
		t.Fatalf("error inesperado en verificación: %v", err)
	}
	if !strings.Contains(res.Nombres, "JUAN CARLOS") {
		t.Errorf("nombres devueltos incorrectos: %s", res.Nombres)
	}

	// Comprobar que se guardó en Redis para autorizar el paso 2
	esVerif, _ := cache.EsVerificado(ctx, "1710034065")
	if !esVerif {
		t.Errorf("esperaba que la cédula quedara marcada como verificada en caché")
	}

	// 2. Tres intentos fallidos generan bloqueo 429
	_ = cache.BorrarVerificado(ctx, "1710034065")
	for i := 1; i <= 3; i++ {
		_, err = svc.Verificar(ctx, VerificarRequest{
			Cedula:         "1710034065",
			CodigoDactilar: &dactilarInvalido,
		})
		if i < 3 {
			var clientErr ClientError
			if !errors.As(err, &clientErr) || clientErr.StatusCode() != 401 {
				t.Fatalf("intento %d: esperaba 401, obtuvo %v", i, err)
			}
		} else {
			var clientErr ClientError
			if !errors.As(err, &clientErr) || clientErr.StatusCode() != 429 {
				t.Fatalf("intento 3: esperaba 429 tras bloqueo, obtuvo %v", err)
			}
		}
	}
}

func TestAuthService_RegistrarYLogin(t *testing.T) {
	svc, cache, _ := setupAuthService()
	ctx := context.Background()

	// 1. Intento de registro sin haber verificado -> 400
	err := svc.Registrar(ctx, RegistrarRequest{
		Cedula:   "1710034065",
		Email:    "juan@example.com",
		Password: "Password123",
	})
	if err == nil {
		t.Fatal("esperaba error al registrarse sin verificación previa")
	}

	// 2. Verificación exitosa previa
	dactilar := "V4443V4444"
	_, _ = svc.Verificar(ctx, VerificarRequest{
		Cedula:         "1710034065",
		CodigoDactilar: &dactilar,
	})

	// 3. Registro con contraseña inválida -> 400
	err = svc.Registrar(ctx, RegistrarRequest{
		Cedula:   "1710034065",
		Email:    "juan@example.com",
		Password: "corta",
	})
	var clientErr ClientError
	if !errors.As(err, &clientErr) || clientErr.StatusCode() != 400 {
		t.Fatalf("esperaba error 400 por contraseña corta, obtuvo %v", err)
	}

	// 4. Registro exitoso -> 201
	err = svc.Registrar(ctx, RegistrarRequest{
		Cedula:   "1710034065",
		Email:    "juan@example.com",
		Password: "Password123",
	})
	if err != nil {
		t.Fatalf("error inesperado en registro: %v", err)
	}

	// Comprobar que se limpió la verificación de Redis
	if verif, _ := cache.EsVerificado(ctx, "1710034065"); verif {
		t.Errorf("esperaba que la verificación se eliminara tras el registro exitoso")
	}

	// 5. Login con contraseña incorrecta -> 401
	_, err = svc.Login(ctx, LoginRequest{
		Cedula:   "1710034065",
		Password: "PasswordIncorrecta",
	})
	if !errors.As(err, &clientErr) || clientErr.StatusCode() != 401 {
		t.Fatalf("esperaba 401 por contraseña incorrecta, obtuvo %v", err)
	}

	// 6. Login exitoso -> 200 con JWT
	loginRes, err := svc.Login(ctx, LoginRequest{
		Cedula:   "1710034065",
		Password: "Password123",
	})
	if err != nil {
		t.Fatalf("error inesperado en login: %v", err)
	}
	if loginRes.Token == "" {
		t.Fatal("esperaba token JWT en la respuesta de login")
	}
}

func TestAuthService_RecuperarYRestablecer(t *testing.T) {
	svc, cache, postRepo := setupAuthService()
	ctx := context.Background()

	// Crear usuario activo
	hash, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	_ = postRepo.Crear(ctx, &model.Postulante{
		Cedula:       "1710034065",
		Email:        "juan@example.com",
		PasswordHash: string(hash),
		RolID:        1,
	})

	dactilar := "V4443V4444"
	// 1. Solicitar recuperación
	err := svc.Recuperar(ctx, RecuperarRequest{
		Cedula:         "1710034065",
		CodigoDactilar: &dactilar,
	})
	if err != nil {
		t.Fatalf("error inesperado al recuperar: %v", err)
	}

	// Obtener el token guardado en cache
	var resetToken string
	for k := range cache.resetTokens {
		resetToken = k
		break
	}
	if resetToken == "" {
		t.Fatal("esperaba token de recuperación guardado en caché")
	}

	// 2. Restablecer con token inválido -> 400
	err = svc.Restablecer(ctx, RestablecerRequest{
		Token:    "token-invalido",
		Password: "NuevaPassword456",
	})
	var clientErr ClientError
	if !errors.As(err, &clientErr) || clientErr.StatusCode() != 400 {
		t.Fatalf("esperaba error 400 por token inválido, obtuvo %v", err)
	}

	// 3. Restablecer con token válido -> 200
	err = svc.Restablecer(ctx, RestablecerRequest{
		Token:    resetToken,
		Password: "NuevaPassword456",
	})
	if err != nil {
		t.Fatalf("error inesperado al restablecer: %v", err)
	}

	// Comprobar que el token se eliminó de la caché (un solo uso, I3)
	if _, ok := cache.resetTokens[resetToken]; ok {
		t.Errorf("el token de recuperación no fue eliminado de la caché")
	}

	// 4. Iniciar sesión con la nueva contraseña
	loginRes, err := svc.Login(ctx, LoginRequest{
		Cedula:   "1710034065",
		Password: "NuevaPassword456",
	})
	if err != nil {
		t.Fatalf("error al iniciar sesión con la nueva contraseña: %v", err)
	}
	if loginRes.Token == "" {
		t.Errorf("token vacío tras login con nueva contraseña")
	}
}
