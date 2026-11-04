package service

import (
	"testing"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
)

func TestFormatosDeCampos(t *testing.T) {
	casos := []struct {
		nombre string
		fn     func(string) bool
		ok     []string
		mal    []string
	}{
		{"EsNombre", EsNombre, []string{"María José Pérez", "Juan O'Neil", "Ana-Lucía Núñez"}, []string{"Juan123", "12345", "A", "Juan@", "<script>"}},
		{"EsTexto", EsTexto, []string{"Universidad Central del Ecuador", "Av. 10 de Agosto N34-120", "Banco 24 de Mayo S.A."}, []string{"12345", "!!!", "<b>Hola</b>", "a=b", "100%"}},
		{"EsNumeroCasa", EsNumeroCasa, []string{"N34-120", "S/N", "sn", "12"}, []string{"casa", "#$%"}},
		{"EsSenescyt", EsSenescyt, []string{"1005-2018-1234567", "ABC123"}, []string{"abc", "12 34", "12/34"}},
		{"EmailValido", EmailValido, []string{"a@example.com"}, []string{"a@b", "a@.com", "Ana <a@b.com>", "a b@c.com", "@c.com"}},
	}
	for _, c := range casos {
		for _, v := range c.ok {
			if !c.fn(v) {
				t.Errorf("%s(%q) = false, esperaba true", c.nombre, v)
			}
		}
		for _, v := range c.mal {
			if c.fn(v) {
				t.Errorf("%s(%q) = true, esperaba false", c.nombre, v)
			}
		}
	}

	if !EsCelular("0991234567", true) || EsCelular("991234567", true) || EsCelular("0291234567", true) || !EsCelular("+573001234567", false) {
		t.Error("EsCelular no respeta el formato de Ecuador / extranjero")
	}
	if !EsTelefonoFijo("022345678", true) || EsTelefonoFijo("0991234567", true) || EsTelefonoFijo("abc", false) {
		t.Error("EsTelefonoFijo no respeta el formato de Ecuador / extranjero")
	}
}

func TestValidarCampos(t *testing.T) {
	s := func(v string) *string { return &v }
	prov, canton := 17, 150
	base := func() *model.PostulacionCompleta {
		return &model.PostulacionCompleta{
			Domicilio:  &model.Domicilio{PaisID: 52, ProvinciaID: &prov, CantonID: &canton, Celular: "0991234567", CallePrincipal: "Av. Amazonas"},
			Familiares: []model.FamiliarIESS{{Nombres: "Ana Pérez", Parentesco: "madre", Institucion: "IESS"}},
		}
	}
	if err := validarCampos(base(), "a@example.com"); err != nil {
		t.Fatalf("postulación válida rechazada: %v", err)
	}

	malos := map[string]func(p *model.PostulacionCompleta){
		"celular con letras":      func(p *model.PostulacionCompleta) { p.Domicilio.Celular = "09912abc67" },
		"celular fijo en Ecuador": func(p *model.PostulacionCompleta) { p.Domicilio.Celular = "022345678" },
		"calle vacía":             func(p *model.PostulacionCompleta) { p.Domicilio.CallePrincipal = " " },
		"calle solo números":      func(p *model.PostulacionCompleta) { p.Domicilio.CallePrincipal = "12345" },
		"email2 igual al 1":       func(p *model.PostulacionCompleta) { p.Domicilio.Email2 = s("A@example.com") },
		"provincia sin cantón":    func(p *model.PostulacionCompleta) { p.Domicilio.CantonID = nil },
		"ciudad extranjera con números": func(p *model.PostulacionCompleta) {
			p.Domicilio.ProvinciaID, p.Domicilio.CantonID = nil, nil
			p.Domicilio.Celular = "+573001234567"
			p.Domicilio.ProvinciaTexto, p.Domicilio.CiudadTexto = s("Antioquia"), s("Medellín 2")
		},
		"familiar con números": func(p *model.PostulacionCompleta) { p.Familiares[0].Nombres = "Ana 123" },
		"capacitación fecha futura": func(p *model.PostulacionCompleta) {
			f := time.Now().AddDate(1, 0, 0)
			p.Capacitaciones = []model.Capacitacion{{NombreEvento: "Curso", Institucion: "IESS", FechaInicio: f, FechaFin: f, Horas: 10}}
		},
		"horas excesivas": func(p *model.PostulacionCompleta) {
			f := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			p.Capacitaciones = []model.Capacitacion{{NombreEvento: "Curso", Institucion: "IESS", FechaInicio: f, FechaFin: f, Horas: 10000}}
		},
	}
	for nombre, cambiar := range malos {
		p := base()
		cambiar(p)
		if err := validarCampos(p, "a@example.com"); err == nil {
			t.Errorf("%s: se aceptó y debía rechazarse", nombre)
		}
	}
}
