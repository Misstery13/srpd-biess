package service

import (
	"fmt"
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

	// Mailpit no requiere autenticación en desarrollo local
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

func formatFechaFin(f *time.Time) string {
	if f == nil {
		return "Actualidad"
	}
	return f.Format("02/01/2006")
}

func renderFormularioHTML(codigo string, p *model.PostulacionCompleta, perfil *model.PostulantePerfil) string {
	var sb strings.Builder
	sb.WriteString("<h2>Formulario de Postulación Registrado</h2>")
	sb.WriteString(fmt.Sprintf("<p><strong>Código de Postulación:</strong> %s</p>", codigo))

	if perfil != nil {
		sb.WriteString("<h3>1. Datos Personales</h3><ul>")
		sb.WriteString(fmt.Sprintf("<li><strong>Cédula:</strong> %s</li>", perfil.Cedula))
		sb.WriteString(fmt.Sprintf("<li><strong>Nombres y Apellidos:</strong> %s %s</li>", perfil.Nombres, perfil.Apellidos))
		sb.WriteString(fmt.Sprintf("<li><strong>Nacionalidad:</strong> %s</li>", perfil.Nacionalidad))
		sb.WriteString(fmt.Sprintf("<li><strong>Estado Civil:</strong> %s</li>", perfil.EstadoCivil))
		sb.WriteString(fmt.Sprintf("<li><strong>Correo Electrónico:</strong> %s</li>", perfil.Email))
		sb.WriteString("</ul>")
	}

	if p != nil {
		if p.Cargo != nil {
			sb.WriteString("<h3>2. Cargo Postulado</h3>")
			sb.WriteString(fmt.Sprintf("<p><strong>Cargo:</strong> %s</p>", *p.Cargo))
			if p.RespaldoOrg != nil && *p.RespaldoOrg && p.OrgNombre != nil {
				sb.WriteString(fmt.Sprintf("<p><strong>Organización de Respaldo:</strong> %s</p>", *p.OrgNombre))
			}
		}

		if p.Domicilio != nil {
			sb.WriteString("<h3>3. Domicilio</h3><ul>")
			sb.WriteString(fmt.Sprintf("<li><strong>Celular:</strong> %s</li>", p.Domicilio.Celular))
			sb.WriteString(fmt.Sprintf("<li><strong>Calle Principal:</strong> %s</li>", p.Domicilio.CallePrincipal))
			if p.Domicilio.CalleSecundaria != nil {
				sb.WriteString(fmt.Sprintf("<li><strong>Calle Secundaria:</strong> %s</li>", *p.Domicilio.CalleSecundaria))
			}
			sb.WriteString("</ul>")
		}

		if len(p.Familiares) > 0 {
			sb.WriteString("<h3>4. Familiares en IESS/BIESS</h3><ul>")
			for _, f := range p.Familiares {
				sb.WriteString(fmt.Sprintf("<li>%s - %s (%s)</li>", f.Nombres, f.Parentesco, f.Institucion))
			}
			sb.WriteString("</ul>")
		}

		if len(p.Formacion) > 0 {
			sb.WriteString("<h3>5. Formación Académica</h3><ul>")
			for _, fa := range p.Formacion {
				sb.WriteString(fmt.Sprintf("<li>%s: %s en %s (%s)</li>", fa.Nivel, fa.Titulo, fa.Institucion, fa.Area))
			}
			sb.WriteString("</ul>")
		}

		if len(p.Experiencia) > 0 {
			sb.WriteString("<h3>6. Experiencia Laboral</h3><ul>")
			for _, el := range p.Experiencia {
				sb.WriteString(fmt.Sprintf("<li>[%s] %s en %s (%s a %s)</li>", el.Tipo, el.Cargo, el.Institucion, el.FechaInicio.Format("02/01/2006"), formatFechaFin(el.FechaFin)))
			}
			sb.WriteString("</ul>")
		}

		if len(p.Publicaciones) > 0 {
			sb.WriteString("<h3>7. Publicaciones</h3><ul>")
			for _, pb := range p.Publicaciones {
				sb.WriteString(fmt.Sprintf("<li>%s: \"%s\" en %s (%s)</li>", pb.Tipo, pb.Titulo, pb.Medio, pb.Fecha.Format("02/01/2006")))
			}
			sb.WriteString("</ul>")
		}

		if len(p.Capacitaciones) > 0 {
			sb.WriteString("<h3>8. Capacitaciones</h3><ul>")
			for _, cp := range p.Capacitaciones {
				sb.WriteString(fmt.Sprintf("<li>%s: %s en %s (%d horas)</li>", cp.TipoEvento, cp.NombreEvento, cp.Institucion, cp.Horas))
			}
			sb.WriteString("</ul>")
		}

		if p.Inhabilidades != nil {
			sb.WriteString("<h3>9. Declaración Jurada de Inhabilidades</h3>")
			sb.WriteString("<p>El postulante declaró bajo juramento no incurrir en inhabilidades legales para el ejercicio de las funciones.</p>")
		}
	}

	sb.WriteString("<p>Atentamente,<br><strong>Comisión Técnica del Concurso BIESS</strong></p>")
	return sb.String()
}
