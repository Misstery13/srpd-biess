package service

import (
	"fmt"
	"log"
	"net/smtp"
)

type EmailService interface {
	EnviarRecuperacion(destinatario, token string) error
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
