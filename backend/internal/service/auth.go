package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/cache"
	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// dummyBcryptHash es un hash bcrypt generado dinámicamente durante la inicialización
// para garantizar tiempo de cómputo constante y mitigar ataques de temporización (timing attacks)
// cuando una cédula/usuario no existe en la base de datos (OWASP A07).
// Se genera en init() para evitar falsos positivos de credenciales fijas en análisis SAST (SonarQube S2068).
var dummyBcryptHash []byte

func init() {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy-timing-protection-seed-constant"), bcrypt.DefaultCost)
	if err == nil {
		dummyBcryptHash = h
	}
}

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

type AuthConfig struct {
	RCRepo    repository.RegistroCivilRepository
	PostRepo  repository.PostulanteRepository
	Cache     cache.AuthCache
	EmailSvc  EmailService
	JWTSecret string
}

func NewAuthService(cfg AuthConfig) AuthService {
	return &authService{
		rcRepo:    cfg.RCRepo,
		postRepo:  cfg.PostRepo,
		cache:     cfg.Cache,
		emailSvc:  cfg.EmailSvc,
		jwtSecret: cfg.JWTSecret,
	}
}

func coincideDactilar(real string, ingresado *string) bool {
	if ingresado == nil || strings.TrimSpace(*ingresado) == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(*ingresado), strings.TrimSpace(real))
}

func coincideFechaExpedicion(real time.Time, ingresada *string) bool {
	if ingresada == nil || strings.TrimSpace(*ingresada) == "" {
		return false
	}
	val := strings.TrimSpace(*ingresada)
	return val == real.Format("2006-01-02") || val == real.Format("02/01/2006")
}

func validarDatosRegistroCivil(rc *model.RegistroCivil, fechaExpedicion, codigoDactilar *string) bool {
	if rc == nil {
		return false
	}
	return coincideDactilar(rc.CodigoDactilar, codigoDactilar) || coincideFechaExpedicion(rc.FechaExpedicion, fechaExpedicion)
}

func (s *authService) verificarBloqueoIntentos(ctx context.Context, tipo, cedula, msgBloqueo string) error {
	bloqueado, err := s.cache.EstaBloqueado(ctx, tipo, cedula)
	if err != nil && err != cache.ErrRedisNoDisponible {
		return NewErrServicioNoDisponible("Servicio temporalmente no disponible")
	}
	if bloqueado {
		log.Printf("[AUTH WARN] Intento con cuenta bloqueada: %s (tipo %s)", EnmascararCedula(cedula), tipo)
		return NewErrBloqueado(msgBloqueo)
	}
	return nil
}

func (s *authService) registrarIntentoFallido(ctx context.Context, tipo, cedula, msgBloqueo string, errRetorno error) error {
	n, _ := s.cache.IncrementarIntentos(ctx, tipo, cedula)
	if n >= cache.MaxIntentos {
		log.Printf("[AUTH ALERTA] Bloqueo aplicado (%s) tras 3 fallos para cédula %s", tipo, EnmascararCedula(cedula))
		return NewErrBloqueado(msgBloqueo)
	}
	return errRetorno
}

func (s *authService) Verificar(ctx context.Context, req VerificarRequest) (*VerificarResponse, error) {
	req.Cedula = strings.TrimSpace(req.Cedula)

	if !ValidarCedula(req.Cedula) {
		return nil, NewErrValidacion("El número de cédula ingresado no es válido")
	}

	if err := s.verificarBloqueoIntentos(ctx, "registro", req.Cedula, "registro fallido, inténtelo nuevamente en 10 minutos"); err != nil {
		return nil, err
	}

	rc, err := s.rcRepo.BuscarPorCedula(ctx, req.Cedula)
	if err != nil {
		return nil, err
	}

	if !validarDatosRegistroCivil(rc, req.FechaExpedicion, req.CodigoDactilar) {
		return nil, s.registrarIntentoFallido(ctx, "registro", req.Cedula, "registro fallido, inténtelo nuevamente en 10 minutos", NewErrNoAutorizado("la información ingresada es incorrecta"))
	}

	existente, err := s.postRepo.BuscarActivoPorCedula(ctx, req.Cedula)
	if err != nil {
		return nil, err
	}
	if existente != nil {
		return nil, NewErrConflicto("La cédula ya se encuentra registrada en el sistema")
	}

	_ = s.cache.GuardarVerificado(ctx, req.Cedula)
	_ = s.cache.ReiniciarIntentos(ctx, "registro", req.Cedula)

	log.Printf("[AUTH] Verificación de identidad exitosa para cédula %s", EnmascararCedula(req.Cedula))
	nombresCompletos := strings.TrimSpace(rc.Nombres + " " + rc.Apellidos)
	return &VerificarResponse{Nombres: nombresCompletos}, nil
}

func validarCamposRegistro(req RegistrarRequest) error {
	if !ValidarCedula(req.Cedula) {
		return NewErrValidacion("El número de cédula ingresado no es válido")
	}
	if !ValidarPassword(req.Password) {
		return NewErrValidacion("La contraseña debe tener al menos 8 caracteres e incluir letras y números")
	}
	if !EmailValido(req.Email) {
		return NewErrValidacion("El correo electrónico no es válido")
	}
	return nil
}

func (s *authService) verificarUnicidadRegistro(ctx context.Context, cedula, email string) error {
	activoCedula, err := s.postRepo.BuscarActivoPorCedula(ctx, cedula)
	if err != nil {
		return err
	}
	if activoCedula != nil {
		return NewErrConflicto("La cédula ya se encuentra registrada en el sistema")
	}

	activoEmail, err := s.postRepo.BuscarActivoPorEmail(ctx, email)
	if err != nil {
		return err
	}
	if activoEmail != nil {
		return NewErrConflicto("El correo electrónico ya se encuentra registrado")
	}
	return nil
}

func (s *authService) obtenerRolInicial(ctx context.Context) (int, error) {
	rol, err := s.postRepo.BuscarRolPorCodigo(ctx, "POSTULANTE")
	if err != nil {
		return 0, err
	}
	if rol != nil {
		return rol.ID, nil
	}
	return 1, nil
}

func (s *authService) Registrar(ctx context.Context, req RegistrarRequest) error {
	req.Cedula = strings.TrimSpace(req.Cedula)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if err := validarCamposRegistro(req); err != nil {
		return err
	}

	verificado, err := s.cache.EsVerificado(ctx, req.Cedula)
	if err != nil && err != cache.ErrRedisNoDisponible {
		return NewErrServicioNoDisponible("Servicio temporalmente no disponible")
	}
	if !verificado {
		return NewErrValidacion("Debe verificar su identidad antes de registrarse")
	}

	if err := s.verificarUnicidadRegistro(ctx, req.Cedula, req.Email); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	rolID, err := s.obtenerRolInicial(ctx)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return NewErrConflicto("La cédula o correo electrónico ya se encuentra registrado")
		}
		return err
	}

	_ = s.cache.BorrarVerificado(ctx, req.Cedula)
	_ = s.cache.ReiniciarIntentos(ctx, "registro", req.Cedula)

	log.Printf("[AUTH] Postulante registrado con éxito: cédula %s", EnmascararCedula(req.Cedula))
	return nil
}

func comprobarCredenciales(postulante *model.Postulante, password string) bool {
	if postulante == nil {
		// Mitigación de Timing Attack: cómputo constante con dummy hash
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(password))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(postulante.PasswordHash), []byte(password)) == nil
}

func generarTokenSesion(postulanteID int, rolCodigo, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.Itoa(postulanteID),
		"rol": rolCodigo,
		"exp": time.Now().Add(2 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (s *authService) resolverCodigoRol(ctx context.Context, rolID int) string {
	if rolID > 0 {
		if rol, err := s.postRepo.BuscarRolPorID(ctx, rolID); err == nil && rol != nil {
			return rol.Codigo
		}
	}
	return "POSTULANTE"
}

func (s *authService) enviarTokenRecuperacion(ctx context.Context, postulante *model.Postulante) {
	if postulante == nil {
		return
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err == nil {
		tokenHex := hex.EncodeToString(tokenBytes)
		_ = s.cache.GuardarResetToken(ctx, tokenHex, postulante.Cedula)
		_ = s.emailSvc.EnviarRecuperacion(postulante.Email, tokenHex)
	}
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	req.Cedula = strings.TrimSpace(req.Cedula)

	if req.Cedula == "" || req.Password == "" {
		return nil, NewErrValidacion("Debe ingresar su cédula y contraseña")
	}

	if err := s.verificarBloqueoIntentos(ctx, "login", req.Cedula, "sus credenciales de ingreso están bloqueadas, inténtelo nuevamente en 10 minutos"); err != nil {
		return nil, err
	}

	// S6: Mensaje genérico de login fallido
	postulante, err := s.postRepo.BuscarActivoPorCedula(ctx, req.Cedula)
	if err != nil {
		return nil, err
	}

	if !comprobarCredenciales(postulante, req.Password) {
		return nil, s.registrarIntentoFallido(ctx, "login", req.Cedula, "sus credenciales de ingreso están bloqueadas, inténtelo nuevamente en 10 minutos", NewErrNoAutorizado("sus credenciales de ingreso son incorrectas"))
	}

	// Éxito: reinicia contador de intentos (S5)
	_ = s.cache.ReiniciarIntentos(ctx, "login", req.Cedula)

	rolCodigo := s.resolverCodigoRol(ctx, postulante.RolID)
	tokenString, err := generarTokenSesion(postulante.ID, rolCodigo, s.jwtSecret)
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

	if err := s.verificarBloqueoIntentos(ctx, "recuperar", req.Cedula, "registro fallido, inténtelo nuevamente en 10 minutos"); err != nil {
		return err
	}

	// Verificar datos contra Registro Civil
	rc, err := s.rcRepo.BuscarPorCedula(ctx, req.Cedula)
	if err != nil {
		return err
	}

	if !validarDatosRegistroCivil(rc, req.FechaExpedicion, req.CodigoDactilar) {
		return s.registrarIntentoFallido(ctx, "recuperar", req.Cedula, "registro fallido, inténtelo nuevamente en 10 minutos", NewErrNoAutorizado("la información ingresada es incorrecta"))
	}

	// Si los datos son correctos, reiniciar intentos de recuperación (S5)
	_ = s.cache.ReiniciarIntentos(ctx, "recuperar", req.Cedula)

	postulante, err := s.postRepo.BuscarActivoPorCedula(ctx, req.Cedula)
	if err != nil {
		return err
	}

	s.enviarTokenRecuperacion(ctx, postulante)
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

	// Reiniciar intentos de login y recuperacion para no dejar bloqueado al postulante
	_ = s.cache.ReiniciarIntentos(ctx, "login", cedula)
	_ = s.cache.ReiniciarIntentos(ctx, "recuperar", cedula)

	log.Printf("[AUTH] Contraseña restablecida con éxito para cédula %s", EnmascararCedula(cedula))
	return nil
}
