package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
)

// Constantes de tamaño para subida de documentos (A3)
const (
	MaxFileSizeOrganizacion = 4 * 1024 * 1024  // 4 MB (A3)
	MaxFileSizeGeneral      = 2 * 1024 * 1024  // 2 MB (A3)
	MaxTotalQuotaPerUser    = 50 * 1024 * 1024 // 50 MB total por postulante (A3)
)

// PostulacionService gestiona las reglas de negocio del formulario EF-12.
type PostulacionService struct {
	postulacionRepo repository.PostulacionRepository
	postulanteRepo  repository.PostulanteRepository
	periodoSvc      *PeriodoService
	emailSvc        EmailService
	uploadsDir      string
}

// NewPostulacionService crea una nueva instancia de PostulacionService.
func NewPostulacionService(
	postulacionRepo repository.PostulacionRepository,
	postulanteRepo repository.PostulanteRepository,
	periodoSvc *PeriodoService,
	emailSvc EmailService,
	uploadsDir string,
) *PostulacionService {
	return &PostulacionService{
		postulacionRepo: postulacionRepo,
		postulanteRepo:  postulanteRepo,
		periodoSvc:      periodoSvc,
		emailSvc:        emailSvc,
		uploadsDir:      uploadsDir,
	}
}

// ValidarCargo valida si el postulante está habilitado según Historial Laboral IESS (1.4.x).
func (s *PostulacionService) ValidarCargo(ctx context.Context, postulanteID int, req *model.ValidarCargoRequest) (*model.ValidarCargoResponse, error) {
	cargo := strings.ToUpper(strings.TrimSpace(req.Cargo))
	if cargo != "AFILIADOS" && cargo != "JUBILADOS" && cargo != "EMPLEADORES" {
		return nil, NewErrValidacion("Cargo inválido. Debe ser AFILIADOS, JUBILADOS o EMPLEADORES.")
	}

	perfil, err := s.postulanteRepo.ObtenerDetalleRegistroCivil(ctx, postulanteID)
	if err != nil {
		return nil, err
	}
	if perfil == nil {
		return nil, NewErrNoEncontrado("Postulante no encontrado.")
	}

	switch cargo {
	case "AFILIADOS":
		// 1.4.1.1 y 1.4.1.2
		if perfil.EsAfiliado {
			return &model.ValidarCargoResponse{
				Habilitado: true,
				Mensaje:    "Postulante afiliado al IESS",
			}, nil
		}
		return &model.ValidarCargoResponse{
			Habilitado: false,
			Mensaje:    "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS",
		}, nil

	case "JUBILADOS":
		// 1.4.2.1 y 1.4.2.2
		if perfil.EsJubilado {
			return &model.ValidarCargoResponse{
				Habilitado: true,
				Mensaje:    "Postulante acredita jubilación",
			}, nil
		}
		return &model.ValidarCargoResponse{
			Habilitado: false,
			Mensaje:    "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación",
		}, nil

	case "EMPLEADORES":
		// 1.4.3.1 y 1.4.3.2
		if perfil.EsEmpleador {
			return &model.ValidarCargoResponse{
				Habilitado: true,
				Mensaje:    "Postulante es empleador",
			}, nil
		}
		return &model.ValidarCargoResponse{
			Habilitado: false,
			Mensaje:    "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador",
		}, nil
	}

	return nil, NewErrValidacion("Cargo desconocido.")
}

// ObtenerPostulacion recupera la postulación completa activa del postulante.
func (s *PostulacionService) ObtenerPostulacion(ctx context.Context, postulanteID int) (*model.PostulacionCompleta, error) {
	p, err := s.postulacionRepo.ObtenerPorPostulanteID(ctx, postulanteID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, NewErrNoEncontrado("No existe postulación registrada para el usuario.")
	}

	// V7: Calcular duración de cada experiencia en años, meses y días
	for i := range p.Experiencia {
		p.Experiencia[i].DuracionTexto = CalcularDuracion(p.Experiencia[i].FechaInicio, p.Experiencia[i].FechaFin)
	}

	return p, nil
}

// GuardarPostulacion valida y persiste la postulación en una única transacción (P4).
func (s *PostulacionService) GuardarPostulacion(ctx context.Context, postulanteID int, p *model.PostulacionCompleta) (string, error) {
	// T2: Validar si el período de postulación está abierto
	periodo, err := s.periodoSvc.ObtenerPeriodo(ctx)
	if err != nil {
		return "", err
	}
	if !periodo.Abierto {
		return "", NewErrProhibido("El período de postulación no está habilitado.")
	}

	// Validaciones de negocio
	if p.Cargo != nil {
		c := strings.ToUpper(strings.TrimSpace(*p.Cargo))
		if c != "AFILIADOS" && c != "JUBILADOS" && c != "EMPLEADORES" {
			return "", NewErrValidacion("El cargo seleccionado no es válido.")
		}
		p.Cargo = &c
	}

	// V7 y V8: Experiencia laboral
	for _, el := range p.Experiencia {
		if el.Funciones != nil && len(*el.Funciones) > 1000 {
			return "", NewErrValidacion("La descripción de funciones no puede exceder los 1000 caracteres (V8).")
		}
		if el.FechaFin != nil && el.FechaFin.Before(el.FechaInicio) {
			return "", NewErrValidacion("La fecha de finalización no puede ser anterior a la fecha de inicio en la experiencia laboral (V7).")
		}
	}

	// Capacitaciones: fecha_fin >= fecha_inicio
	for _, cp := range p.Capacitaciones {
		if cp.FechaFin.Before(cp.FechaInicio) {
			return "", NewErrValidacion("La fecha de finalización no puede ser anterior a la fecha de inicio en capacitación.")
		}
		if cp.Horas <= 0 {
			return "", NewErrValidacion("Las horas de capacitación deben ser un número positivo.")
		}
	}

	// Inhabilidades: validación de campos condicionales obligatorios
	if p.Inhabilidades != nil && p.Inhabilidades.RemovidoControl {
		if p.Inhabilidades.RemovidoInstitucion == nil || strings.TrimSpace(*p.Inhabilidades.RemovidoInstitucion) == "" ||
			p.Inhabilidades.RemovidoMotivo == nil || strings.TrimSpace(*p.Inhabilidades.RemovidoMotivo) == "" {
			return "", NewErrValidacion("Si fue removido por un organismo de control, debe especificar institución y motivo.")
		}
	}

	codigo, err := s.postulacionRepo.GuardarTransaccional(ctx, postulanteID, p)
	if err != nil {
		if err.Error() == "POSTULACION_ENVIADA_NO_EDITABLE" {
			return "", NewErrProhibido("Una postulación ya enviada no puede ser modificada (P3).")
		}
		return "", err
	}

	return codigo, nil
}

// SubirDocumento valida el archivo (%PDF, tamaño, cuota) y lo almacena con nombre UUID (A1-A3, O1).
func (s *PostulacionService) SubirDocumento(ctx context.Context, postulanteID int, tipo string, fileHeader *multipart.FileHeader) (string, error) {
	// T2: Validar período
	periodo, err := s.periodoSvc.ObtenerPeriodo(ctx)
	if err != nil {
		return "", err
	}
	if !periodo.Abierto {
		return "", NewErrProhibido("El período de postulación no está habilitado.")
	}

	// P3: Verificar estado de postulación
	estado, err := s.postulacionRepo.ObtenerEstado(ctx, postulanteID)
	if err != nil {
		return "", err
	}
	if estado == "ENVIADA" || estado == "CERRADA" {
		return "", NewErrProhibido("No se pueden subir documentos a una postulación que ya ha sido enviada (P3).")
	}

	// A3: Límite por tipo de archivo
	tipoUpper := strings.ToUpper(strings.TrimSpace(tipo))
	maxSize := int64(MaxFileSizeGeneral)
	if tipoUpper == "ORGANIZACION" {
		maxSize = MaxFileSizeOrganizacion
	}

	if fileHeader.Size > maxSize {
		limiteMB := maxSize / (1024 * 1024)
		return "", NewErrValidacion(fmt.Sprintf("El archivo excede el tamaño máximo permitido de %d MB.", limiteMB))
	}

	// Abrir y leer los primeros bytes para validar el tipo real (A2)
	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("error al abrir archivo: %w", err)
	}
	defer src.Close()

	headerBytes := make([]byte, 5)
	if _, err := io.ReadFull(src, headerBytes); err != nil {
		return "", NewErrValidacion("El archivo está vacío o dañado.")
	}

	// A2: Validación de número mágico %PDF
	if !bytes.HasPrefix(headerBytes, []byte("%PDF")) {
		return "", NewErrValidacion("El archivo subido no es un documento PDF válido.")
	}

	// Reiniciar puntero de lectura al inicio
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("error al reiniciar lectura de archivo: %w", err)
	}

	// A1, O1: Carpeta de destino uploads/{postulante_id}/
	userUploadsDir := filepath.Join(s.uploadsDir, fmt.Sprintf("%d", postulanteID))
	if err := os.MkdirAll(userUploadsDir, 0750); err != nil {
		return "", fmt.Errorf("error al crear directorio de almacenamiento: %w", err)
	}

	// A3: Validar cuota total acumulada del postulante (máximo 50 MB)
	totalExistente, err := calcularTamanoDirectorio(userUploadsDir)
	if err != nil {
		return "", fmt.Errorf("error al verificar cuota de almacenamiento: %w", err)
	}
	if totalExistente+fileHeader.Size > MaxTotalQuotaPerUser {
		return "", NewErrValidacion("Se ha excedido el límite de almacenamiento total de 50 MB para este postulante.")
	}

	// A1: Nombre UUID aleatorio para evitar path traversal
	uuidNombre := fmt.Sprintf("%s.pdf", generarUUID())
	destPath := filepath.Join(userUploadsDir, uuidNombre)

	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		return "", fmt.Errorf("error al crear archivo en disco: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		_ = os.Remove(destPath)
		return "", fmt.Errorf("error al escribir archivo en disco: %w", err)
	}

	// Retorna la ruta relativa para persistencia en base de datos
	rutaRelativa := filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("%d", postulanteID), uuidNombre))
	return rutaRelativa, nil
}

// EnviarPostulacion finaliza y envía la postulación con consentimiento explícito LOPDP (L1, P3, I2).
func (s *PostulacionService) EnviarPostulacion(ctx context.Context, postulanteID int, req *model.EnviarPostulacionRequest) (string, error) {
	// T2: Validar período
	periodo, err := s.periodoSvc.ObtenerPeriodo(ctx)
	if err != nil {
		return "", err
	}
	if !periodo.Abierto {
		return "", NewErrProhibido("El período de postulación no está habilitado.")
	}

	// L1: Consentimiento explícito obligatorio
	if req == nil || !req.AceptaTerminos {
		return "", NewErrValidacion("Debe aceptar los términos y condiciones de consentimiento para enviar la postulación.")
	}

	perfil, err := s.postulanteRepo.ObtenerDetalleRegistroCivil(ctx, postulanteID)
	if err != nil {
		return "", err
	}
	if perfil == nil {
		return "", NewErrNoEncontrado("Postulante no encontrado.")
	}

	codigo, err := s.postulacionRepo.EnviarPostulacion(ctx, postulanteID)
	if err != nil {
		switch err.Error() {
		case "POSTULACION_NO_ENCONTRADA":
			return "", NewErrNoEncontrado("No existe una postulación activa para enviar.")
		case "POSTULACION_YA_ENVIADA":
			return "", NewErrConflicto("La postulación ya ha sido enviada anteriormente.")
		case "DEBE_COMPLETAR_DOMICILIO":
			return "", NewErrValidacion("Debe completar la sección de domicilio antes de enviar la postulación.")
		case "DEBE_SELECCIONAR_CARGO":
			return "", NewErrValidacion("Debe seleccionar el cargo al que postula antes de enviar.")
		case "DEBE_DECLARAR_INHABILIDADES":
			return "", NewErrValidacion("Debe responder la declaración jurada de inhabilidades antes de enviar.")
		default:
			return "", err
		}
	}

	// I2: Envío de correo asíncrono a Mailpit
	_ = s.emailSvc.EnviarConfirmacionPostulacion(perfil.Email, codigo)

	return codigo, nil
}

// Helper para calcular duración en años, meses y días (V7)
func CalcularDuracion(inicio time.Time, fin *time.Time) string {
	hasta := time.Now()
	if fin != nil {
		hasta = *fin
	}
	if hasta.Before(inicio) {
		return "0 días"
	}

	years := hasta.Year() - inicio.Year()
	months := int(hasta.Month()) - int(inicio.Month())
	days := hasta.Day() - inicio.Day()

	if days < 0 {
		prevMonth := hasta.AddDate(0, -1, 0)
		daysInPrevMonth := time.Date(prevMonth.Year(), prevMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		days += daysInPrevMonth
		months--
	}
	if months < 0 {
		months += 12
		years--
	}

	var parts []string
	if years > 0 {
		if years == 1 {
			parts = append(parts, "1 año")
		} else {
			parts = append(parts, fmt.Sprintf("%d años", years))
		}
	}
	if months > 0 {
		if months == 1 {
			parts = append(parts, "1 mes")
		} else {
			parts = append(parts, fmt.Sprintf("%d meses", months))
		}
	}
	if days > 0 || len(parts) == 0 {
		if days == 1 {
			parts = append(parts, "1 día")
		} else {
			parts = append(parts, fmt.Sprintf("%d días", days))
		}
	}
	return strings.Join(parts, ", ")
}

func calcularTamanoDirectorio(dir string) (int64, error) {
	var total int64
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func generarUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // Versión 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variante 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
