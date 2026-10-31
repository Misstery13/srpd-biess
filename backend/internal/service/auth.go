package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/cache"
	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type VerificarRequest struct {
	Cedula          string  `json:"cedula"`
	FechaExpedicion *string `json:"fecha_expedicion,omitempty"`
	CodigoDactilar  *string `json:"codigo_dactilar,omitempty"`
}

type VerificarResponse struct {
	Nombres string `json:"nombres"`
}

type RegistrarRequest struct {
	Cedula   string `json:"cedula"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Cedula   string `json:"cedula"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RecuperarRequest struct {
	Cedula          string  `json:"cedula"`
	FechaExpedicion *string `json:"fecha_expedicion,omitempty"`
	CodigoDactilar  *string `json:"codigo_dactilar,omitempty"`
}

type RestablecerRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type AuthService interface {
	Verificar(ctx context.Context, req VerificarRequest) (*VerificarResponse, error)
	Registrar(ctx context.Context, req RegistrarRequest) error
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Recuperar(ctx context.Context, req RecuperarRequest) error
	Restablecer(ctx context.Context, req RestablecerRequest) error
}

type authService struct {
	rcRepo    repository.RegistroCivilRepository
	postRepo  repository.PostulanteRepository
	cache     cache.AuthCache
	emailSvc  EmailService
	jwtSecret string
}

func NewAuthService(
	rcRepo repository.RegistroCivilRepository,
	postRepo repository.PostulanteRepository,
	cache cache.AuthCache,
	emailSvc EmailService,
	jwtSecret string,
) AuthService {
	return &authService{
		rcRepo:    rcRepo,
		postRepo:  postRepo,
		cache:     cache,
		emailSvc:  emailSvc,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Verificar(ctx context.Context, req VerificarRequest) (*VerificarResponse, error) {
	req.Cedula = strings.TrimSpace(req.Cedula)

	// V1: Validación de cédula ecuatoriana módulo 10
	if !ValidarCedula(req.Cedula) {
		return nil, NewErrValidacion("El número de cédula ingresado no es válido")
	}

	// S4: Verificar si está bloqueado por 3 intentos fallidos
	bloqueado, err := s.cache.EstaBloqueado(ctx, "registro", req.Cedula)
	if err != nil && err != cache.ErrRedisNoDisponible {
		return nil, NewErrServicioNoDisponible("Servicio temporalmente no disponible")
	}
	if bloqueado {
		log.Printf("[AUTH WARN] Intento de verificación con cédula bloqueada: %s", EnmascararCedula(req.Cedula))
		return nil, NewErrBloqueado("registro fallido, inténtelo nuevamente en 10 minutos")
	}

	// Consultar interoperabilidad simulada con Registro Civil (I1)
	rc, err := s.rcRepo.BuscarPorCedula(ctx, req.Cedula)
	if err != nil {
		return nil, err
	}

	datosValidos := false
	if rc != nil {
		// Validar fecha de expedición o código dactilar
		if req.CodigoDactilar != nil && strings.TrimSpace(*req.CodigoDactilar) != "" {
			dactilarIngresado := strings.ToUpper(strings.TrimSpace(*req.CodigoDactilar))
			dactilarReal := strings.ToUpper(strings.TrimSpace(rc.CodigoDactilar))
			if dactilarIngresado == dactilarReal {
				datosValidos = true
			}
		}

		if !datosValidos && req.FechaExpedicion != nil && strings.TrimSpace(*req.FechaExpedicion) != "" {
			fechaIngresada := strings.TrimSpace(*req.FechaExpedicion)
			fechaISO := rc.FechaExpedicion.Format("2006-01-02")
			fechaUI := rc.FechaExpedicion.Format("02/01/2006")
			if fechaIngresada == fechaISO || fechaIngresada == fechaUI {
				datosValidos = true
			}
		}
	}

	if !datosValidos {
		// Incrementar intentos y bloquear si llega a 3 (S4)
		n, _ := s.cache.IncrementarIntentos(ctx, "registro", req.Cedula)
		if n >= cache.MaxIntentos {
			log.Printf("[AUTH ALERTA] Bloqueo aplicado tras 3 intentos fallidos en verificación para cédula %s", EnmascararCedula(req.Cedula))
			return nil, NewErrBloqueado("registro fallido, inténtelo nuevamente en 10 minutos")
		}
		return nil, NewErrNoAutorizado("la información ingresada es incorrecta")
	}

	// Verificar si la cédula ya tiene cuenta activa registrada
	existente, err := s.postRepo.BuscarActivoPorCedula(ctx, req.Cedula)
	if err != nil {
		return nil, err
	}
	if existente != nil {
		return nil, NewErrConflicto("La cédula ya se encuentra registrada en el sistema")
	}

	// Guardar estado verificado en Redis para autorizar el paso 2 (TTL 10 min)
	_ = s.cache.GuardarVerificado(ctx, req.Cedula)
	_ = s.cache.ReiniciarIntentos(ctx, "registro", req.Cedula)

	log.Printf("[AUTH] Verificación de identidad exitosa para cédula %s", EnmascararCedula(req.Cedula))
	nombresCompletos := strings.TrimSpace(rc.Nombres + " " + rc.Apellidos)
	return &VerificarResponse{Nombres: nombresCompletos}, nil
}

func (s *authService) Registrar(ctx context.Context, req RegistrarRequest) error {
	req.Cedula = strings.TrimSpace(req.Cedula)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if !ValidarCedula(req.Cedula) {
		return NewErrValidacion("El número de cédula ingresado no es válido")
	}

	// Exigir paso 1 previo en Redis
	verificado, err := s.cache.EsVerificado(ctx, req.Cedula)
	if err != nil && err != cache.ErrRedisNoDisponible {
		return NewErrServicioNoDisponible("Servicio temporalmente no disponible")
	}
	if !verificado {
		return NewErrValidacion("Debe verificar su identidad antes de registrarse")
	}

	// S3: Contraseña mínima 8 caracteres, al menos 1 letra y 1 número
	if !ValidarPassword(req.Password) {
		return NewErrValidacion("La contraseña debe tener al menos 8 caracteres e incluir letras y números")
	}

	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return NewErrValidacion("El correo electrónico no es válido")
	}

	// Verificar unicidad de cédula y email entre activos (V4, E2)
	activoPorCedula, err := s.postRepo.BuscarActivoPorCedula(ctx, req.Cedula)
	if err != nil {
		return err
	}
	if activoPorCedula != nil {
		return NewErrConflicto("La cédula ya se encuentra registrada en el sistema")
	}

	activoPorEmail, err := s.postRepo.BuscarActivoPorEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if activoPorEmail != nil {
		return NewErrConflicto("El correo electrónico ya se encuentra registrado")
	}

	// Rol inicial POSTULANTE (R1, R2)
	rol, err := s.postRepo.BuscarRolPorCodigo(ctx, "POSTULANTE")
	if err != nil {
		return err
	}
	rolID := 1
	if rol != nil {
		rolID = rol.ID
	}

	// S1: Hash con bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	postulante := &model.Postulante{
		Cedula:       req.Cedula,
		Email:        req.Email,
		PasswordHash: string(hash),
		RolID:        rolID,
	}

	if err := s.postRepo.Crear(ctx, postulante); err != nil {
		return err
	}

	// Limpiar caché de verificación y reintentos (S5)
	_ = s.cache.BorrarVerificado(ctx, req.Cedula)
	_ = s.cache.ReiniciarIntentos(ctx, "registro", req.Cedula)

	log.Printf("[AUTH] Postulante registrado con éxito: cédula %s", EnmascararCedula(req.Cedula))
	return nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	req.Cedula = strings.TrimSpace(req.Cedula)

	if req.Cedula == "" || req.Password == "" {
		return nil, NewErrValidacion("Debe ingresar su cédula y contraseña")
	}

	// S4: Verificar si está bloqueado por 3 intentos
	bloqueado, err := s.cache.EstaBloqueado(ctx, "login", req.Cedula)
	if err != nil && err != cache.ErrRedisNoDisponible {
		return nil, NewErrServicioNoDisponible("Servicio temporalmente no disponible")
	}
	if bloqueado {
		log.Printf("[AUTH WARN] Intento de login con cuenta bloqueada: %s", EnmascararCedula(req.Cedula))
		return nil, NewErrBloqueado("sus credenciales de ingreso están bloqueadas, inténtelo nuevamente en 10 minutos")
	}

	// S6: Mensaje genérico de login fallido
	postulante, err := s.postRepo.BuscarActivoPorCedula(ctx, req.Cedula)
	if err != nil {
		return nil, err
	}

	credencialesInvalidas := false
	if postulante == nil {
		credencialesInvalidas = true
	} else if err := bcrypt.CompareHashAndPassword([]byte(postulante.PasswordHash), []byte(req.Password)); err != nil {
		credencialesInvalidas = true
	}

	if credencialesInvalidas {
		n, _ := s.cache.IncrementarIntentos(ctx, "login", req.Cedula)
		if n >= cache.MaxIntentos {
			log.Printf("[AUTH ALERTA] Bloqueo de login aplicado para cédula %s", EnmascararCedula(req.Cedula))
			return nil, NewErrBloqueado("sus credenciales de ingreso están bloqueadas, inténtelo nuevamente en 10 minutos")
		}
		return nil, NewErrNoAutorizado("sus credenciales de ingreso son incorrectas")
	}

	// Éxito: reinicia contador de intentos (S5)
	_ = s.cache.ReiniciarIntentos(ctx, "login", req.Cedula)

	// S2, R2: JWT de 2 horas con sub y rol
	claims := jwt.MapClaims{
		"sub":    strconv.Itoa(postulante.ID),
		"rol":    "POSTULANTE",
		"cedula": postulante.Cedula,
		"exp":    time.Now().Add(2 * time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	log.Printf("[AUTH] Login exitoso para cédula %s", EnmascararCedula(req.Cedula))
	return &LoginResponse{Token: tokenString}, nil
}

func (s *authService) Recuperar(ctx context.Context, req RecuperarRequest) error {
	req.Cedula = strings.TrimSpace(req.Cedula)

	if !ValidarCedula(req.Cedula) {
		return NewErrValidacion("El número de cédula ingresado no es válido")
	}

	// Verificar datos contra Registro Civil
	rc, err := s.rcRepo.BuscarPorCedula(ctx, req.Cedula)
	if err != nil {
		return err
	}

	datosValidos := false
	if rc != nil {
		if req.CodigoDactilar != nil && strings.TrimSpace(*req.CodigoDactilar) != "" {
			if strings.ToUpper(strings.TrimSpace(*req.CodigoDactilar)) == strings.ToUpper(strings.TrimSpace(rc.CodigoDactilar)) {
				datosValidos = true
			}
		}

		if !datosValidos && req.FechaExpedicion != nil && strings.TrimSpace(*req.FechaExpedicion) != "" {
			fechaIngresada := strings.TrimSpace(*req.FechaExpedicion)
			if fechaIngresada == rc.FechaExpedicion.Format("2006-01-02") || fechaIngresada == rc.FechaExpedicion.Format("02/01/2006") {
				datosValidos = true
			}
		}
	}

	if !datosValidos {
		return NewErrNoAutorizado("la información ingresada es incorrecta")
	}

	// Buscar si está registrado como postulante activo
	postulante, err := s.postRepo.BuscarActivoPorCedula(ctx, req.Cedula)
	if err != nil {
		return err
	}

	if postulante != nil {
		// I3: Generar token de recuperación aleatorio (crypto/rand) con TTL 15 min
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err == nil {
			tokenHex := hex.EncodeToString(tokenBytes)
			_ = s.cache.GuardarResetToken(ctx, tokenHex, postulante.Cedula)
			_ = s.emailSvc.EnviarRecuperacion(postulante.Email, tokenHex)
		}
	}

	log.Printf("[AUTH] Solicitud de recuperación procesada para cédula %s", EnmascararCedula(req.Cedula))
	return nil
}

func (s *authService) Restablecer(ctx context.Context, req RestablecerRequest) error {
	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		return NewErrValidacion("El enlace de recuperación es inválido o ha expirado")
	}

	if !ValidarPassword(req.Password) {
		return NewErrValidacion("La contraseña debe tener al menos 8 caracteres e incluir letras y números")
	}

	// I3: Token de un solo uso en Redis
	cedula, err := s.cache.ObtenerYBorrarResetToken(ctx, req.Token)
	if err != nil && err != cache.ErrRedisNoDisponible {
		return NewErrServicioNoDisponible("Servicio temporalmente no disponible")
	}
	if cedula == "" {
		return NewErrValidacion("El enlace de recuperación es inválido o ha expirado")
	}

	// Actualizar contraseña con bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.postRepo.ActualizarPassword(ctx, cedula, string(hash)); err != nil {
		return err
	}

	log.Printf("[AUTH] Contraseña restablecida con éxito para cédula %s", EnmascararCedula(cedula))
	return nil
}
