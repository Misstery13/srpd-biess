package service

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
)

// Formatos de los campos de EF-12 (mismas reglas en frontend/src/utils/validaciones.js)
var (
	reNombre     = regexp.MustCompile(`^[\p{L}\p{M}' .-]+$`)
	reTexto      = regexp.MustCompile(`^[\p{L}\p{M}\p{N}\s.,;:()'"&/#°º+¿?¡!@_–-]+$`)
	reNumeroCasa = regexp.MustCompile(`^[\p{L}\p{N} /.-]+$`)
	reSinNumero  = regexp.MustCompile(`(?i)^s/?n$`)
	reSenescyt   = regexp.MustCompile(`^[A-Za-z0-9-]+$`)
	reCelularEC  = regexp.MustCompile(`^09\d{8}$`)
	reFijoEC     = regexp.MustCompile(`^0[2-7]\d{7}$`)
	reTelefonoEx = regexp.MustCompile(`^\+?\d{7,15}$`)
	reDigito     = regexp.MustCompile(`\d`)
	fechaMinima  = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
)

const MaxHorasCapacitacion = 9999

func contarLetras(s string) int {
	n := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			n++
		}
	}
	return n
}

func EsNombre(s string) bool { return reNombre.MatchString(s) && contarLetras(s) >= 2 }
func EsTexto(s string) bool  { return reTexto.MatchString(s) && contarLetras(s) >= 1 }
func EsNumeroCasa(s string) bool {
	return reSinNumero.MatchString(s) || (reNumeroCasa.MatchString(s) && reDigito.MatchString(s))
}
func EsSenescyt(s string) bool { return reSenescyt.MatchString(s) && reDigito.MatchString(s) }

func EsCelular(s string, ecuador bool) bool {
	if ecuador {
		return reCelularEC.MatchString(s)
	}
	return reTelefonoEx.MatchString(s)
}

func EsTelefonoFijo(s string, ecuador bool) bool {
	if ecuador {
		return reFijoEC.MatchString(s)
	}
	return reTelefonoEx.MatchString(s)
}

// EmailValido: dirección simple (sin nombre visible) con dominio que tenga punto
func EmailValido(s string) bool {
	a, err := mail.ParseAddress(s)
	if err != nil || a.Address != s {
		return false
	}
	dominio := s[strings.LastIndex(s, "@")+1:]
	return strings.Contains(dominio, ".") && !strings.HasPrefix(dominio, ".") && !strings.HasSuffix(dominio, ".")
}

// fechaEnRango: entre 1900-01-01 y hoy (fecha de Ecuador)
func fechaEnRango(f time.Time) bool {
	n := time.Now().In(zonaEcuador)
	hoy := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
	dia := time.Date(f.Year(), f.Month(), f.Day(), 0, 0, 0, 0, time.UTC) // solo la fecha, sin hora
	return !dia.Before(fechaMinima) && !dia.After(hoy)
}

// revisor acumula el primer error de validación encontrado
type revisor struct{ err error }

func (r *revisor) falla(msg string) {
	if r.err == nil {
		r.err = NewErrValidacion(msg)
	}
}

// texto valida un campo obligatorio (obligatorio=true) u opcional con formato de texto o nombre
func (r *revisor) texto(v *string, etiqueta string, obligatorio, esNombre bool) {
	s := ""
	if v != nil {
		s = strings.TrimSpace(*v)
	}
	switch {
	case s == "":
		if obligatorio {
			r.falla("El campo " + etiqueta + " es obligatorio.")
		}
	case esNombre && !EsNombre(s):
		r.falla("El campo " + etiqueta + " solo puede contener letras.")
	case !esNombre && !EsTexto(s):
		r.falla("El campo " + etiqueta + " debe contener letras y no puede tener símbolos como < > { } $ % * =.")
	}
}

func (r *revisor) fecha(f *time.Time, etiqueta string) {
	if f != nil && !fechaEnRango(*f) {
		r.falla("La " + etiqueta + " debe estar entre 1900 y hoy.")
	}
}

// validarCampos revisa obligatorios y formato de todos los campos de texto, teléfonos y fechas (1.2–1.9)
func validarCampos(p *model.PostulacionCompleta, emailRegistrado string) error {
	r := &revisor{}
	if p.RespaldoOrg != nil && *p.RespaldoOrg {
		r.texto(p.OrgNombre, "nombre de organización", true, false)
	}

	if d := p.Domicilio; d != nil {
		ecuador := d.ProvinciaID != nil
		if ecuador && d.CantonID == nil {
			r.falla("Seleccione la ciudad del domicilio.")
		}
		if !ecuador {
			r.texto(d.ProvinciaTexto, "provincia / estado", true, true)
			r.texto(d.CiudadTexto, "ciudad", true, true)
		}
		cel := strings.TrimSpace(d.Celular)
		if cel == "" {
			r.falla("El campo teléfono celular es obligatorio.")
		} else if !EsCelular(cel, ecuador) {
			r.falla("El teléfono celular no es válido (Ecuador: 10 dígitos que empiezan con 09).")
		}
		if d.Telefono != nil && strings.TrimSpace(*d.Telefono) != "" && !EsTelefonoFijo(strings.TrimSpace(*d.Telefono), ecuador) {
			r.falla("El teléfono de domicilio no es válido (Ecuador: 9 dígitos con el código de provincia).")
		}
		if d.Email2 != nil && strings.TrimSpace(*d.Email2) != "" {
			e2 := strings.TrimSpace(*d.Email2)
			if !EmailValido(e2) {
				r.falla("El correo electrónico 2 no es válido.")
			} else if strings.EqualFold(e2, emailRegistrado) {
				r.falla("El correo electrónico 2 debe ser distinto del correo registrado.")
			}
		}
		r.texto(&d.CallePrincipal, "calle principal", true, false)
		if d.Numero != nil && strings.TrimSpace(*d.Numero) != "" && !EsNumeroCasa(strings.TrimSpace(*d.Numero)) {
			r.falla("El número de domicilio no es válido (use números y letras, p. ej. N34-120, o S/N).")
		}
		r.texto(d.CalleSecundaria, "calle secundaria", false, false)
		r.texto(d.Sector, "sector", false, false)
		r.texto(d.Referencia, "referencia", false, false)
	}

	for i := range p.Familiares {
		f := &p.Familiares[i]
		r.texto(&f.Nombres, "nombres y apellidos del familiar", true, true)
		r.texto(f.Area, "área del familiar", false, false)
		r.texto(f.Cargo, "cargo del familiar", false, false)
	}
	for i := range p.Formacion {
		f := &p.Formacion[i]
		r.texto(&f.Institucion, "institución educativa", true, false)
		r.texto(&f.Titulo, "título obtenido", true, false)
		if f.SenescytNumero != nil && strings.TrimSpace(*f.SenescytNumero) != "" && !EsSenescyt(strings.TrimSpace(*f.SenescytNumero)) {
			r.falla("El número de registro SENESCYT solo admite números, letras y guiones.")
		}
		r.fecha(f.SenescytFecha, "fecha de registro SENESCYT")
	}
	for i := range p.Experiencia {
		e := &p.Experiencia[i]
		r.texto(&e.Cargo, "cargo o puesto", true, false)
		r.texto(&e.Institucion, "institución o empresa", true, false)
		r.texto(e.Funciones, "descripción de funciones", false, false)
		r.fecha(&e.FechaInicio, "fecha de inicio de la experiencia")
		r.fecha(e.FechaFin, "fecha de finalización de la experiencia")
	}
	for i := range p.Publicaciones {
		pb := &p.Publicaciones[i]
		r.texto(&pb.Titulo, "título de la publicación", true, false)
		r.texto(&pb.Medio, "medio de publicación", true, false)
		r.fecha(&pb.Fecha, "fecha de publicación")
	}
	for i := range p.Capacitaciones {
		c := &p.Capacitaciones[i]
		r.texto(&c.NombreEvento, "nombre del evento", true, false)
		r.texto(&c.Institucion, "institución de la capacitación", true, false)
		r.fecha(&c.FechaInicio, "fecha de inicio de la capacitación")
		r.fecha(&c.FechaFin, "fecha de fin de la capacitación")
		if c.Horas > MaxHorasCapacitacion {
			r.falla("Las horas de capacitación no pueden superar 9999.")
		}
	}
	if in := p.Inhabilidades; in != nil && in.RemovidoControl != nil && *in.RemovidoControl {
		r.texto(in.RemovidoInstitucion, "institución (removido)", true, false)
		r.texto(in.RemovidoCargo, "cargo (removido)", false, false)
		r.texto(in.RemovidoMotivo, "motivo de finalización", true, false)
		r.fecha(in.RemovidoFechaFin, "fecha de finalización (removido)")
	}
	return r.err
}
