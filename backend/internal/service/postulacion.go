package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

var rutaDocRegex = regexp.MustCompile(`^uploads/(\d+)/[0-9a-f-]{36}\.pdf$`)

// Listas fijas según el requerimiento funcional EF-12
var parentescosValidos = map[string]bool{
	"conviviente": true, "cónyuge": true, "conyuge": true, "cuñado/a": true, "cunado/a": true,
	"hermano/a": true, "hijo político": true, "hijo politico": true, "hijo/a": true,
	"madre": true, "nieto/a": true, "nuera": true, "padre": true, "primo hermano/a": true,
	"sobrino/a": true, "suegro/a": true, "tío/a": true, "tio/a": true, "yerno": true,
}

var nivelesInstruccionValidos = map[string]bool{
	"tercer nivel de grado":                 true,
	"cuarto nivel - especialidad académica": true,
	"cuarto nivel - especialidad academica": true,
	"cuarto nivel - maestría académica":     true,
	"cuarto nivel - maestria academica":     true,
	"cuarto nivel phd - doctorado":          true,
}

var tiposFormacionValidos = map[string]string{
	"nacional":   "NACIONAL",
	"extranjera": "EXTRANJERA",
}

var areasEstudioValidas = map[string]bool{
	"finanzas":                   true,
	"mercado de capitales":       true,
	"administración de empresas": true,
	"administracion de empresas": true,
	"economía":                   true,
	"economia":                   true,
	"derecho o materias afines":  true,
}

var tiposExperienciaValidos = map[string]string{
	"general":   "GENERAL",
	"dirección": "DIRECCION",
	"direccion": "DIRECCION",
}

var areasExperienciaValidas = map[string]bool{
	"finanzas":                   true,
	"mercado de capitales":       true,
	"administración de empresas": true,
	"administracion de empresas": true,
	"economía":                   true,
	"economia":                   true,
	"derecho o materias afines":  true,
}

var tiposPublicacionValidos = map[string]string{
	"publicación":              "PUBLICACION",
	"publicacion":              "PUBLICACION",
	"trabajo de investigación": "INVESTIGACION",
	"trabajo de investigacion": "INVESTIGACION",
	"investigacion":            "INVESTIGACION",
	"investigación":            "INVESTIGACION",
}

var tiposEventoCapacitacionValidos = map[string]string{
	"curso":     "CURSO",
	"seminario": "SEMINARIO",
	"taller":    "TALLER",
}

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

// postulanteHabilitado valida nacionalidad ecuatoriana (V5, 1.1.6), cuenta activa (L5) y requisitos de cargo (1.4.x).
func (s *PostulacionService) postulanteHabilitado(ctx context.Context, id int, cargo *string) error {
	perfil, err := s.postulanteRepo.ObtenerDetalleRegistroCivil(ctx, id) // ya filtra deleted_at IS NULL
	if err != nil {
		return err
	}
	if perfil == nil {
		return NewErrNoAutorizado("Su sesión ya no es válida.") // cuenta suprimida → el frontend cierra sesión (E3)
	}
	if perfil.Nacionalidad != "ECUATORIANA" {
		return NewErrProhibido("Solo ciudadanos con nacionalidad ecuatoriana pueden postular.")
	}
	if cargo != nil && *cargo != "" {
		if r, err := s.ValidarCargo(ctx, id, &model.ValidarCargoRequest{Cargo: *cargo}); err != nil {
			return err
		} else if r != nil && !r.Habilitado {
			return NewErrProhibido(r.Mensaje)
		}
	}
	return nil
}

// validarRutaDoc verifica que el documento cumpla el patrón seguro y pertenezca al usuario (O1).
func validarRutaDoc(ruta *string, postulanteID int, campo string) error {
	if ruta == nil || strings.TrimSpace(*ruta) == "" {
		return nil
	}
	m := rutaDocRegex.FindStringSubmatch(strings.TrimSpace(*ruta))
	if m == nil || m[1] != strconv.Itoa(postulanteID) {
		return NewErrValidacion(fmt.Sprintf("La ruta de documento en %s no es válida o no pertenece al postulante.", campo))
	}
	return nil
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
	if err := s.postulanteHabilitado(ctx, postulanteID, nil); err != nil {
		return nil, err
	}

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
	// 1, 2, 3: Validar que el postulante esté habilitado y apto para el cargo
	if err := s.postulanteHabilitado(ctx, postulanteID, p.Cargo); err != nil {
		return "", err
	}

	// T2: Validar si el período de postulación está abierto
	periodo, err := s.periodoSvc.ObtenerPeriodo(ctx)
	if err != nil {
		return "", err
	}
	if !periodo.Abierto {
		return "", NewErrProhibido("El período de postulación no está habilitado.")
	}

	// Validaciones de negocio del cargo
	if p.Cargo != nil {
		c := strings.ToUpper(strings.TrimSpace(*p.Cargo))
		if c != "AFILIADOS" && c != "JUBILADOS" && c != "EMPLEADORES" {
			return "", NewErrValidacion("El cargo seleccionado no es válido.")
		}
		p.Cargo = &c
	}

	// 6: Campos obligatorios de respaldo organizacional (1.4.x.3.1.1 y 1.4.x.3.1.2)
	if p.RespaldoOrg != nil && *p.RespaldoOrg {
		if p.OrgNombre == nil || strings.TrimSpace(*p.OrgNombre) == "" {
			return "", NewErrValidacion("El nombre de la organización es obligatorio cuando se indica respaldo (1.4.x.3.1.1).")
		}
		if p.OrgDocumento == nil || strings.TrimSpace(*p.OrgDocumento) == "" {
			return "", NewErrValidacion("El documento de respaldo es obligatorio cuando se indica respaldo (1.4.x.3.1.2).")
		}
	}
	if err := validarRutaDoc(p.OrgDocumento, postulanteID, "organización"); err != nil {
		return "", err
	}

	// 5: Validar listas fijas de familiares (1.3.1.2 y 1.3.1.3)
	for _, f := range p.Familiares {
		parentescoNorm := strings.ToLower(strings.TrimSpace(f.Parentesco))
		if !parentescosValidos[parentescoNorm] {
			return "", NewErrValidacion(fmt.Sprintf("El parentesco '%s' no es válido según la lista oficial (1.3.1.2).", f.Parentesco))
		}
		instNorm := strings.ToUpper(strings.TrimSpace(f.Institucion))
		if instNorm != "IESS" && instNorm != "BIESS" {
			return "", NewErrValidacion(fmt.Sprintf("La institución familiar '%s' debe ser IESS o BIESS (1.3.1.3).", f.Institucion))
		}
	}

	// 5: Validar listas fijas de formación académica (1.5.1, 1.5.1.2, 1.5.1.4)
	for i := range p.Formacion {
		fa := &p.Formacion[i]
		nivelNorm := strings.ToLower(strings.TrimSpace(fa.Nivel))
		if !nivelesInstruccionValidos[nivelNorm] {
			return "", NewErrValidacion(fmt.Sprintf("El nivel de instrucción '%s' no es válido según la lista oficial (1.5.1).", fa.Nivel))
		}
		tipoNorm := strings.ToLower(strings.TrimSpace(fa.Tipo))
		dbTipo, ok := tiposFormacionValidos[tipoNorm]
		if !ok {
			return "", NewErrValidacion(fmt.Sprintf("El tipo de institución '%s' no es válido. Debe ser Nacional o Extranjera (1.5.1.2).", fa.Tipo))
		}
		fa.Tipo = dbTipo

		areaNorm := strings.ToLower(strings.TrimSpace(fa.Area))
		if !areasEstudioValidas[areaNorm] {
			return "", NewErrValidacion(fmt.Sprintf("El área de estudios '%s' no es válida según la lista oficial (1.5.1.4).", fa.Area))
		}

		if err := validarRutaDoc(fa.Documento, postulanteID, "formación académica"); err != nil {
			return "", err
		}
	}

	// 5 y 6: Validar experiencia laboral (1.6.1, 1.6.6, 1.6.8, V7, V8)
	for i := range p.Experiencia {
		el := &p.Experiencia[i]
		tipoNorm := strings.ToLower(strings.TrimSpace(el.Tipo))
		dbTipo, ok := tiposExperienciaValidos[tipoNorm]
		if !ok {
			return "", NewErrValidacion(fmt.Sprintf("El tipo de experiencia '%s' no es válido. Debe ser General o Dirección (1.6.1).", el.Tipo))
		}
		el.Tipo = dbTipo

		areaNorm := strings.ToLower(strings.TrimSpace(el.Area))
		if !areasExperienciaValidas[areaNorm] {
			return "", NewErrValidacion(fmt.Sprintf("El área de actividades '%s' no es válida según la lista oficial (1.6.6).", el.Area))
		}

		if el.Documento == nil || strings.TrimSpace(*el.Documento) == "" {
			return "", NewErrValidacion("El documento de respaldo es obligatorio para cada registro de experiencia laboral (1.6.8).")
		}
		if err := validarRutaDoc(el.Documento, postulanteID, "experiencia laboral"); err != nil {
			return "", err
		}

		if el.Funciones != nil && len(*el.Funciones) > 1000 {
			return "", NewErrValidacion("La descripción de funciones no puede exceder los 1000 caracteres (V8).")
		}
		if el.FechaFin != nil && el.FechaFin.Before(el.FechaInicio) {
			return "", NewErrValidacion("La fecha de finalización no puede ser anterior a la fecha de inicio en la experiencia laboral (V7).")
		}
	}

	// 5 y 6: Validar publicaciones (1.7.1.2, 1.7.1.6)
	for i := range p.Publicaciones {
		pb := &p.Publicaciones[i]
		tipoNorm := strings.ToLower(strings.TrimSpace(pb.Tipo))
		dbTipo, ok := tiposPublicacionValidos[tipoNorm]
		if !ok {
			return "", NewErrValidacion(fmt.Sprintf("El tipo de publicación '%s' no es válido. Debe ser Publicación o Trabajo de investigación (1.7.1.2).", pb.Tipo))
		}
		pb.Tipo = dbTipo

		if pb.Documento == nil || strings.TrimSpace(*pb.Documento) == "" {
			return "", NewErrValidacion("El documento de respaldo es obligatorio para cada publicación o trabajo de investigación (1.7.1.6).")
		}
		if err := validarRutaDoc(pb.Documento, postulanteID, "publicaciones"); err != nil {
			return "", err
		}
	}

	// 5: Validar capacitaciones (1.8.1)
	for i := range p.Capacitaciones {
		cp := &p.Capacitaciones[i]
		tipoNorm := strings.ToLower(strings.TrimSpace(cp.TipoEvento))
		dbTipo, ok := tiposEventoCapacitacionValidos[tipoNorm]
		if !ok {
			return "", NewErrValidacion(fmt.Sprintf("El tipo de evento '%s' no es válido. Debe ser Curso, Seminario o Taller (1.8.1).", cp.TipoEvento))
		}
		cp.TipoEvento = dbTipo

		if cp.FechaFin.Before(cp.FechaInicio) {
			return "", NewErrValidacion("La fecha de finalización no puede ser anterior a la fecha de inicio en capacitación.")
		}
		if cp.Horas <= 0 {
			return "", NewErrValidacion("Las horas de capacitación deben ser un número positivo.")
		}
		if err := validarRutaDoc(cp.Documento, postulanteID, "capacitaciones"); err != nil {
			return "", err
		}
	}

	// Inhabilidades: validación de campos condicionales obligatorios
	if p.Inhabilidades != nil && p.Inhabilidades.RemovidoControl {
		if p.Inhabilidades.RemovidoInstitucion == nil || strings.TrimSpace(*p.Inhabilidades.RemovidoInstitucion) == "" ||
			p.Inhabilidades.RemovidoMotivo == nil || strings.TrimSpace(*p.Inhabilidades.RemovidoMotivo) == "" {
			return "", NewErrValidacion("Si fue removido por un organismo de control, debe especificar institución y motivo.")
		}
	}

	// Obtener postulación previa para identificar documentos reemplazados
	oldPostulacion, _ := s.postulacionRepo.ObtenerPorPostulanteID(ctx, postulanteID)

	codigo, err := s.postulacionRepo.GuardarTransaccional(ctx, postulanteID, p)
	if err != nil {
		if errors.Is(err, repository.ErrPostulacionEnviadaNoEditable) || err.Error() == "POSTULACION_ENVIADA_NO_EDITABLE" {
			return "", NewErrProhibido("Una postulación ya enviada no puede ser modificada (P3).")
		}
		return "", err
	}

	// Limpiar archivos físicos en disco que fueron reemplazados para liberar cuota de almacenamiento
	if oldPostulacion != nil && s.uploadsDir != "" {
		s.limpiarArchivosHuerfanos(postulanteID, oldPostulacion, p)
	}

	return codigo, nil
}

func (s *PostulacionService) limpiarArchivosHuerfanos(postulanteID int, oldP, newP *model.PostulacionCompleta) {
	nuevosDocs := make(map[string]bool)
	agregarDoc := func(d *string) {
		if d != nil && strings.TrimSpace(*d) != "" {
			nuevosDocs[strings.TrimSpace(*d)] = true
		}
	}
	agregarDoc(newP.OrgDocumento)
	for _, fa := range newP.Formacion {
		agregarDoc(fa.Documento)
	}
	for _, el := range newP.Experiencia {
		agregarDoc(el.Documento)
	}
	for _, pb := range newP.Publicaciones {
		agregarDoc(pb.Documento)
	}
	for _, cp := range newP.Capacitaciones {
		agregarDoc(cp.Documento)
	}

	chequearDoc := func(d *string) {
		if d != nil && strings.TrimSpace(*d) != "" && !nuevosDocs[strings.TrimSpace(*d)] {
			m := rutaDocRegex.FindStringSubmatch(strings.TrimSpace(*d))
			if m != nil && m[1] == strconv.Itoa(postulanteID) {
				rel := strings.TrimPrefix(strings.TrimSpace(*d), "uploads/")
				filePath := filepath.Join(s.uploadsDir, rel)
				_ = os.Remove(filePath)
			}
		}
	}

	chequearDoc(oldP.OrgDocumento)
	for _, fa := range oldP.Formacion {
		chequearDoc(fa.Documento)
	}
	for _, el := range oldP.Experiencia {
		chequearDoc(el.Documento)
	}
	for _, pb := range oldP.Publicaciones {
		chequearDoc(pb.Documento)
	}
	for _, cp := range oldP.Capacitaciones {
		chequearDoc(cp.Documento)
	}
}

// SubirDocumento valida el archivo (%PDF, tamaño, cuota) y lo almacena con nombre UUID (A1-A3, O1).
func (s *PostulacionService) SubirDocumento(ctx context.Context, postulanteID int, tipo string, fileHeader *multipart.FileHeader) (string, error) {
	// 1, 3: Validar que el postulante esté habilitado
	if err := s.postulanteHabilitado(ctx, postulanteID, nil); err != nil {
		return "", err
	}

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

	// Recuperar la postulación actual guardada en base de datos
	postulacion, err := s.postulacionRepo.ObtenerPorPostulanteID(ctx, postulanteID)
	if err != nil {
		return "", err
	}
	if postulacion == nil {
		return "", NewErrNoEncontrado("No existe una postulación activa para enviar.")
	}

	// 1, 2, 3: Validar que el postulante esté habilitado y cumpla el cargo guardado en BD
	if err := s.postulanteHabilitado(ctx, postulanteID, postulacion.Cargo); err != nil {
		return "", err
	}

	perfil, err := s.postulanteRepo.ObtenerDetalleRegistroCivil(ctx, postulanteID)
	if err != nil {
		return "", err
	}
	if perfil == nil {
		return "", NewErrNoAutorizado("Su sesión ya no es válida.")
	}

	codigo, err := s.postulacionRepo.EnviarPostulacion(ctx, postulanteID)
	if err != nil {
		if errors.Is(err, repository.ErrPostulacionNoEncontrada) || err.Error() == "POSTULACION_NO_ENCONTRADA" {
			return "", NewErrNoEncontrado("No existe una postulación activa para enviar.")
		}
		if errors.Is(err, repository.ErrPostulacionYaEnviada) || err.Error() == "POSTULACION_YA_ENVIADA" {
			return "", NewErrConflicto("La postulación ya ha sido enviada anteriormente.")
		}
		if errors.Is(err, repository.ErrDebeCompletarDomicilio) || err.Error() == "DEBE_COMPLETAR_DOMICILIO" {
			return "", NewErrValidacion("Debe completar la sección de domicilio antes de enviar la postulación.")
		}
		if errors.Is(err, repository.ErrDebeSeleccionarCargo) || err.Error() == "DEBE_SELECCIONAR_CARGO" {
			return "", NewErrValidacion("Debe seleccionar el cargo al que postula antes de enviar.")
		}
		if errors.Is(err, repository.ErrDebeDeclararInhabilidades) || err.Error() == "DEBE_DECLARAR_INHABILIDADES" {
			return "", NewErrValidacion("Debe responder la declaración jurada de inhabilidades antes de enviar.")
		}
		return "", err
	}

	// Envío síncrono de correo a Mailpit con formulario completo (1.10.3, I2)
	_ = s.emailSvc.EnviarConfirmacionPostulacion(perfil.Email, codigo, postulacion, perfil)

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
