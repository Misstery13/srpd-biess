package service

import (
	"fmt"
	"html"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
)

type EmailService interface {
	EnviarRecuperacion(destinatario, token string) error
	EnviarConfirmacionPostulacion(destinatario, codigo string, postulacion *model.PostulacionCompleta, perfil *model.PostulantePerfil) error
}

type smtpEmailService struct {
	host   string
	port   string
	appURL string
}

func NewEmailService(host, port, appURL string) EmailService {
	return &smtpEmailService{
		host:   host,
		port:   port,
		appURL: appURL,
	}
}

func (s *smtpEmailService) EnviarRecuperacion(destinatario, token string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	from := "noreply@biess.gob.ec"
	subject := "Recuperación de credenciales - SRPD BIESS"

	link := fmt.Sprintf("%s/restablecer?token=%s", s.appURL, token)

	body := fmt.Sprintf("Subject: %s\r\n"+
		"From: %s\r\n"+
		"To: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		"<p>Estimado/a postulante,</p>"+
		"<p>Ha solicitado recuperar el acceso a su cuenta en el Concurso de Méritos y Oposición BIESS.</p>"+
		"<p>Para restablecer su contraseña, haga clic en el siguiente enlace (válido por 15 minutos):</p>"+
		"<p><a href=\"%s\">Restablecer contraseña</a></p>"+
		"<p>Si usted no solicitó este cambio, puede ignorar este mensaje.</p>",
		subject, from, destinatario, link,
	)

	err := smtp.SendMail(addr, nil, from, []string{destinatario}, []byte(body))
	if err != nil {
		log.Printf("[EMAIL WARN] No se pudo enviar correo de recuperación: %v", err)
		return err
	}

	log.Printf("[EMAIL] Correo de recuperación enviado satisfactoriamente a destinatario")
	return nil
}

func (s *smtpEmailService) EnviarConfirmacionPostulacion(destinatario, codigo string, postulacion *model.PostulacionCompleta, perfil *model.PostulantePerfil) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	from := "noreply@biess.gob.ec"
	subject := fmt.Sprintf("Confirmación de Envío de Postulación - %s", codigo)

	contenidoHTML := renderFormularioHTML(codigo, postulacion, perfil)

	body := fmt.Sprintf("Subject: %s\r\n"+
		"From: %s\r\n"+
		"To: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		"<p>Estimado/a postulante,</p>"+
		"<p>Su postulación para el Concurso de Méritos y Oposición del Directorio del BIESS ha sido enviada exitosamente.</p>"+
		"<hr>"+
		"%s"+
		"<hr>"+
		"<p>Guarde este mensaje y su código para futuras consultas y el seguimiento del concurso.</p>",
		subject, from, destinatario, contenidoHTML,
	)

	err := smtp.SendMail(addr, nil, from, []string{destinatario}, []byte(body))
	if err != nil {
		log.Printf("[EMAIL WARN] No se pudo enviar confirmacion de postulacion: %v", err)
		return err
	}

	// I2, L4: El log solo dice "correo enviado a postulación BIESS-2026-XXXXXX", nunca datos personales (L4/O8)
	log.Printf("[EMAIL] correo enviado a postulación %s", codigo)
	return nil
}

func esc(s string) string {
	return html.EscapeString(s)
}

func escPtr(s *string, fallback string) string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return fallback
	}
	return html.EscapeString(strings.TrimSpace(*s))
}

func siNo(b *bool) string {
	if b == nil {
		return "No respondido"
	}
	if *b {
		return "Sí"
	}
	return "No"
}

func siNoVal(b bool) string {
	if b {
		return "Sí"
	}
	return "No"
}

func formatFecha(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("02/01/2006")
}

func formatFechaFin(f *time.Time) string {
	if f == nil {
		return "Actualidad"
	}
	return f.Format("02/01/2006")
}

func cargoLegible(c *string) string {
	if c == nil {
		return "No especificado"
	}
	switch strings.ToUpper(strings.TrimSpace(*c)) {
	case "AFILIADOS":
		return "Miembro Principal y Alterno en Representación del Sector de los Afiliados"
	case "JUBILADOS":
		return "Miembro Principal y Alterno en Representación del Sector de los Jubilados"
	case "EMPLEADORES":
		return "Miembro Principal y Alterno en Representación del Sector Empleador"
	default:
		return esc(*c)
	}
}

func tipoFormacionLegible(t string) string {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case "NACIONAL":
		return "Nacional"
	case "EXTRANJERA":
		return "Extranjera"
	default:
		return esc(t)
	}
}

func tipoExperienciaLegible(t string) string {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case "GENERAL":
		return "General"
	case "DIRECCION":
		return "Dirección"
	default:
		return esc(t)
	}
}

func tipoPublicacionLegible(t string) string {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case "PUBLICACION":
		return "Publicación"
	case "INVESTIGACION":
		return "Trabajo de investigación"
	default:
		return esc(t)
	}
}

func tipoEventoLegible(t string) string {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case "CURSO":
		return "Curso"
	case "SEMINARIO":
		return "Seminario"
	case "TALLER":
		return "Taller"
	default:
		return esc(t)
	}
}

func renderFormularioHTML(codigo string, p *model.PostulacionCompleta, perfil *model.PostulantePerfil) string {
	var sb strings.Builder
	sb.WriteString("<h2>Formulario de Postulación Registrado</h2>\n")
	sb.WriteString(fmt.Sprintf("<p><strong>Código de Postulación:</strong> %s</p>\n", esc(codigo)))

	// 1. Información Personal
	if perfil != nil {
		sb.WriteString("<h3>1. Datos Personales</h3>\n<ul>\n")
		sb.WriteString(fmt.Sprintf("<li><strong>Cédula:</strong> %s</li>\n", esc(perfil.Cedula)))
		sb.WriteString(fmt.Sprintf("<li><strong>Nombres y Apellidos:</strong> %s %s</li>\n", esc(perfil.Nombres), esc(perfil.Apellidos)))
		sb.WriteString(fmt.Sprintf("<li><strong>Género:</strong> %s</li>\n", esc(perfil.Genero)))
		sb.WriteString(fmt.Sprintf("<li><strong>Fecha de Nacimiento:</strong> %s</li>\n", formatFecha(perfil.FechaNacimiento)))
		sb.WriteString(fmt.Sprintf("<li><strong>Nacionalidad:</strong> %s</li>\n", esc(perfil.Nacionalidad)))
		sb.WriteString(fmt.Sprintf("<li><strong>Estado Civil:</strong> %s</li>\n", esc(perfil.EstadoCivil)))
		if perfil.ConyugeNombres != nil && strings.TrimSpace(*perfil.ConyugeNombres) != "" {
			sb.WriteString(fmt.Sprintf("<li><strong>Cónyuge / Conviviente:</strong> %s (Cédula: %s)</li>\n",
				esc(*perfil.ConyugeNombres), escPtr(perfil.ConyugeCedula, "N/A")))
		}
		sb.WriteString(fmt.Sprintf("<li><strong>Correo Electrónico Principal:</strong> %s</li>\n", esc(perfil.Email)))
		sb.WriteString("</ul>\n")
	}

	if p != nil {
		// 2. Cargo Postulado
		sb.WriteString("<h3>2. Cargo Postulado</h3>\n<ul>\n")
		sb.WriteString(fmt.Sprintf("<li><strong>Cargo:</strong> %s</li>\n", cargoLegible(p.Cargo)))
		respaldoTexto := "No"
		if p.RespaldoOrg != nil && *p.RespaldoOrg {
			respaldoTexto = "Sí"
		}
		sb.WriteString(fmt.Sprintf("<li><strong>Respaldo de Organización:</strong> %s</li>\n", respaldoTexto))
		if p.RespaldoOrg != nil && *p.RespaldoOrg {
			sb.WriteString(fmt.Sprintf("<li><strong>Nombre de Organización:</strong> %s</li>\n", escPtr(p.OrgNombre, "N/A")))
			sb.WriteString(fmt.Sprintf("<li><strong>Documento de Respaldo:</strong> %s</li>\n", adjunto(p.OrgDocumento)))
		}
		sb.WriteString("</ul>\n")

		// 3. Domicilio
		if p.Domicilio != nil {
			d := p.Domicilio
			sb.WriteString("<h3>3. Domicilio</h3>\n<ul>\n")

			pais := "Ecuador"
			if d.PaisNombre != nil && *d.PaisNombre != "" {
				pais = *d.PaisNombre
			} else if d.PaisID != 52 && d.PaisID != 0 {
				pais = fmt.Sprintf("País #%d", d.PaisID)
			}
			sb.WriteString(fmt.Sprintf("<li><strong>País:</strong> %s</li>\n", esc(pais)))

			prov := escPtr(d.ProvinciaNombre, "")
			if prov == "" {
				prov = escPtr(d.ProvinciaTexto, "N/A")
			}
			sb.WriteString(fmt.Sprintf("<li><strong>Provincia:</strong> %s</li>\n", prov))

			canton := escPtr(d.CantonNombre, "")
			if canton == "" {
				canton = escPtr(d.CiudadTexto, "N/A")
			}
			sb.WriteString(fmt.Sprintf("<li><strong>Cantón / Ciudad:</strong> %s</li>\n", canton))

			sb.WriteString(fmt.Sprintf("<li><strong>Calle Principal:</strong> %s</li>\n", esc(d.CallePrincipal)))
			sb.WriteString(fmt.Sprintf("<li><strong>Número:</strong> %s</li>\n", escPtr(d.Numero, "S/N")))
			sb.WriteString(fmt.Sprintf("<li><strong>Calle Secundaria:</strong> %s</li>\n", escPtr(d.CalleSecundaria, "N/A")))
			sb.WriteString(fmt.Sprintf("<li><strong>Sector:</strong> %s</li>\n", escPtr(d.Sector, "N/A")))
			sb.WriteString(fmt.Sprintf("<li><strong>Referencia:</strong> %s</li>\n", escPtr(d.Referencia, "N/A")))
			sb.WriteString(fmt.Sprintf("<li><strong>Celular:</strong> %s</li>\n", esc(d.Celular)))
			sb.WriteString(fmt.Sprintf("<li><strong>Teléfono Convencional:</strong> %s</li>\n", escPtr(d.Telefono, "N/A")))
			sb.WriteString(fmt.Sprintf("<li><strong>Correo Electrónico Alternativo:</strong> %s</li>\n", escPtr(d.Email2, "N/A")))
			sb.WriteString("</ul>\n")
		}

		// 4. Familiares en IESS/BIESS
		sb.WriteString("<h3>4. Información Adicional - Familiares en IESS/BIESS</h3>\n")
		if len(p.Familiares) > 0 {
			sb.WriteString("<ul>\n")
			for _, f := range p.Familiares {
				sb.WriteString(fmt.Sprintf("<li><strong>%s</strong> — Parentesco: %s | Institución: %s | Área: %s | Cargo: %s</li>\n",
					esc(f.Nombres), esc(f.Parentesco), esc(f.Institucion), escPtr(f.Area, "N/A"), escPtr(f.Cargo, "N/A")))
			}
			sb.WriteString("</ul>\n")
		} else {
			sb.WriteString("<p>No registra familiares laborando en el IESS o BIESS.</p>\n")
		}

		// 5. Formación Académica
		sb.WriteString("<h3>5. Formación Académica</h3>\n")
		if len(p.Formacion) > 0 {
			sb.WriteString("<ul>\n")
			for _, fa := range p.Formacion {
				senFecha := "N/A"
				if fa.SenescytFecha != nil {
					senFecha = formatFecha(*fa.SenescytFecha)
				}
				sb.WriteString(fmt.Sprintf("<li><strong>%s</strong>: %s — %s (%s) | Área: %s | Reg. SENESCYT: %s (Fecha: %s) | Respaldo: %s</li>\n",
					esc(fa.Nivel), esc(fa.Titulo), esc(fa.Institucion), tipoFormacionLegible(fa.Tipo), esc(fa.Area),
					escPtr(fa.SenescytNumero, "N/A"), senFecha, adjunto(fa.Documento)))
			}
			sb.WriteString("</ul>\n")
		} else {
			sb.WriteString("<p>No registra títulos académicos.</p>\n")
		}

		// 6. Experiencia Laboral
		sb.WriteString("<h3>6. Experiencia Laboral</h3>\n")
		if len(p.Experiencia) > 0 {
			sb.WriteString("<ul>\n")
			for _, el := range p.Experiencia {
				duracion := el.DuracionTexto
				if duracion == "" {
					duracion = CalcularDuracion(el.FechaInicio, el.FechaFin)
				}
				sb.WriteString(fmt.Sprintf("<li><strong>[%s] %s</strong> en %s | Período: %s a %s (Duración: %s) | Área: %s<br>Funciones: %s<br>Respaldo: %s</li>\n",
					tipoExperienciaLegible(el.Tipo), esc(el.Cargo), esc(el.Institucion),
					formatFecha(el.FechaInicio), formatFechaFin(el.FechaFin), esc(duracion), esc(el.Area),
					escPtr(el.Funciones, "N/A"), adjunto(el.Documento)))
			}
			sb.WriteString("</ul>\n")
		} else {
			sb.WriteString("<p>No registra experiencia laboral.</p>\n")
		}

		// 7. Publicaciones
		sb.WriteString("<h3>7. Publicaciones y Trabajos de Investigación</h3>\n")
		if len(p.Publicaciones) > 0 {
			sb.WriteString("<ul>\n")
			for _, pb := range p.Publicaciones {
				sb.WriteString(fmt.Sprintf("<li><strong>%s</strong>: &ldquo;%s&rdquo; | Medio: %s | Fecha: %s | Relación con el cargo: %s | Respaldo: %s</li>\n",
					tipoPublicacionLegible(pb.Tipo), esc(pb.Titulo), esc(pb.Medio), formatFecha(pb.Fecha),
					siNoVal(pb.RelacionProfesional), adjunto(pb.Documento)))
			}
			sb.WriteString("</ul>\n")
		} else {
			sb.WriteString("<p>No registra publicaciones ni trabajos de investigación.</p>\n")
		}

		// 8. Capacitaciones
		sb.WriteString("<h3>8. Capacitaciones</h3>\n")
		if len(p.Capacitaciones) > 0 {
			sb.WriteString("<ul>\n")
			for _, cp := range p.Capacitaciones {
				sb.WriteString(fmt.Sprintf("<li><strong>%s: %s</strong> — %s | Período: %s al %s | Horas: %d | Respaldo: %s</li>\n",
					tipoEventoLegible(cp.TipoEvento), esc(cp.NombreEvento), esc(cp.Institucion),
					formatFecha(cp.FechaInicio), formatFecha(cp.FechaFin), cp.Horas, adjunto(cp.Documento)))
			}
			sb.WriteString("</ul>\n")
		} else {
			sb.WriteString("<p>No registra eventos de capacitación.</p>\n")
		}

		// 9. Inhabilidades (Declaración juramentada detallada de las 13 preguntas)
		sb.WriteString("<h3>9. Declaración Jurada de Inhabilidades</h3>\n")
		if p.Inhabilidades != nil {
			inh := p.Inhabilidades
			sb.WriteString("<ol>\n")
			sb.WriteString(fmt.Sprintf("<li>¿Es Ecuatoriano/a y está en pleno goce de los derechos de participación política?: <strong>%s</strong></li>\n", siNo(inh.GoceDerechosPoliticos)))
			sb.WriteString(fmt.Sprintf("<li>¿Se encuentra inhabilitado para ejercer el comercio?: <strong>%s</strong></li>\n", siNo(inh.InhabilitadoComercio)))
			sb.WriteString(fmt.Sprintf("<li>¿Está en mora, directa o indirectamente, en el pago de sus obligaciones en cualquiera de las instituciones del Estado o de las instituciones bajo control de la Superintendencia de Bancos, incluidas las off-shore?: <strong>%s</strong></li>\n", siNo(inh.MoraObligacionesEstado)))
			sb.WriteString(fmt.Sprintf("<li>¿Mantiene algún vínculo, por propiedad o administración, con las instituciones del sistema financiero y del sistema de seguros privados, con las bolsas de valores del país, Casas de Valores y Administradoras de Fondos y Fideicomisos, de acuerdo con las normas que la Junta de Política y Regulación Monetaria y Financiera emita para el efecto?: <strong>%s</strong></li>\n", siNo(inh.VinculoSistemaFinanciero)))
			sb.WriteString(fmt.Sprintf("<li>¿Es funcionario o empleado del IESS - BIESS?: <strong>%s</strong></li>\n", siNo(inh.FuncionarioIessBiess)))
			sb.WriteString(fmt.Sprintf("<li>¿Mantiene interés propio o representa a terceros en la propiedad, la dirección o la gestión de las compañías aseguradoras u otras personas que integran el sistema nacional de seguridad social?: <strong>%s</strong></li>\n", siNo(inh.InteresAseguradoras)))

			removidoResp := siNo(inh.RemovidoControl)
			if inh.RemovidoControl != nil && *inh.RemovidoControl {
				fechaRem := "N/A"
				if inh.RemovidoFechaFin != nil {
					fechaRem = formatFecha(*inh.RemovidoFechaFin)
				}
				removidoResp += fmt.Sprintf("<br>&nbsp;&nbsp;&bull; Institución: %s<br>&nbsp;&nbsp;&bull; Cargo: %s<br>&nbsp;&nbsp;&bull; Fecha: %s<br>&nbsp;&nbsp;&bull; Motivo: %s",
					escPtr(inh.RemovidoInstitucion, "N/A"), escPtr(inh.RemovidoCargo, "N/A"), fechaRem, escPtr(inh.RemovidoMotivo, "N/A"))
			}
			sb.WriteString(fmt.Sprintf("<li>¿En los últimos 5 años, ha sido removido por algún Organismo de Control?: <strong>%s</strong></li>\n", removidoResp))

			sb.WriteString(fmt.Sprintf("<li>¿Mantiene obligaciones pendientes con el SRI?: <strong>%s</strong></li>\n", siNo(inh.ObligacionesSRI)))
			sb.WriteString(fmt.Sprintf("<li>¿Ha incurrido en los últimos 5 años en castigo de sus obligaciones por parte de cualquier institución financiera?: <strong>%s</strong></li>\n", siNo(inh.CastigoObligacionesFinancieras)))
			sb.WriteString(fmt.Sprintf("<li>¿Actualmente se encuentra litigando en contra del IESS - BIESS?: <strong>%s</strong></li>\n", siNo(inh.LitigioIessBiess)))
			sb.WriteString(fmt.Sprintf("<li>¿Se encuentra procesado, enjuiciado o condenado por algún delito relacionado con corrupción o crimen organizado?: <strong>%s</strong></li>\n", siNo(inh.ProcesadoCorrupcion)))
			sb.WriteString(fmt.Sprintf("<li>¿Mantiene en firme algún tipo de responsabilidades por la Contraloría General del Estado?: <strong>%s</strong></li>\n", siNo(inh.ResponsabilidadContraloria)))
			sb.WriteString(fmt.Sprintf("<li>¿Se encuentra registrado en la base de datos de la Unidad de Análisis Financiero y Económico?: <strong>%s</strong></li>\n", siNo(inh.RegistradoUAFE)))
			sb.WriteString("</ol>\n")
		} else {
			sb.WriteString("<p>No registra respuestas de inhabilidades.</p>\n")
		}

		// 10. Consentimientos y Declaraciones (1.10.1)
		sb.WriteString("<h3>10. Consentimientos y Declaraciones Aceptadas (1.10.1)</h3>\n")
		sb.WriteString("<ol>\n")
		sb.WriteString("<li>Autorizo de manera libre y voluntaria, emitiendo consentimiento explícito al Instituto Ecuatoriano de Seguridad Social, para el uso de mis datos personales dentro del concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores, de conformidad con lo prescrito en el artículo 7 de la Ley Orgánica de Protección de Datos Personales. [ACEPTADO]</li>\n")
		sb.WriteString("<li>Declaro que la información del presente formulario y documentación adjunta corresponde a la verdad y autorizo al Instituto Ecuatoriano de Seguridad Social para que se compruebe ante cualquier organismo público o privado su veracidad. [ACEPTADO]</li>\n")
		sb.WriteString("<li>Acepto expresamente cumplir con las normas aplicables al concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores; así como las resoluciones y disposiciones impartidas por el Instituto Ecuatoriano de Seguridad Social. En caso de incumplimiento de requisitos o encontrarme inmerso/a en las prohibiciones o inhabilidades acepto mi descalificación del proceso. [ACEPTADO]</li>\n")
		sb.WriteString("<li>Acepto que mi nombre como postulante al concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores; sea publicado en la página web del IESS - BIESS. [ACEPTADO]</li>\n")
		sb.WriteString("<li>El presente formulario, debe ser impreso y con firma manuscrita, entregado junto con los requisitos solicitados, en el lugar, fecha y horario establecido por el Instituto Ecuatoriano de Seguridad Social. [ACEPTADO]</li>\n")
		sb.WriteString("</ol>\n")

		fechaConsentimiento := time.Now().In(zonaEcuador).Format("02/01/2006 15:04:05")
		if p.FechaConsentimiento != nil {
			fechaConsentimiento = p.FechaConsentimiento.In(zonaEcuador).Format("02/01/2006 15:04:05")
		}
		sb.WriteString(fmt.Sprintf("<p><strong>Fecha y hora del consentimiento:</strong> %s</p>\n", fechaConsentimiento))
	}

	sb.WriteString("<p>Atentamente,<br><strong>Comisión Técnica del Concurso BIESS</strong></p>")
	return sb.String()
}

// zonaEcuador: hora continental de Ecuador (UTC-5, sin horario de verano); FixedZone no depende de tzdata en la imagen
var zonaEcuador = time.FixedZone("ECT", -5*60*60)

// adjunto no expone la ruta interna del archivo en el correo
func adjunto(d *string) string {
	if d == nil || strings.TrimSpace(*d) == "" {
		return "No adjunto"
	}
	return "Documento cargado"
}
