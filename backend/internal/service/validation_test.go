package service

import (
	"testing"
)

func TestValidarCedula(t *testing.T) {
	tests := []struct {
		name   string
		cedula string
		valida bool
	}{
		{"Cédula válida Juan Pérez", "1710034065", true},
		{"Cédula válida Ana Mora", "0912345675", true},
		{"Cédula válida Pedro Salazar", "0102030400", true},
		{"Cédula válida Carlos Gómez", "1754321980", true},
		{"Cédula muy corta", "171003406", false},
		{"Cédula muy larga", "17100340655", false},
		{"Cédula con letras", "171003406A", false},
		{"Provincia inválida (00)", "0010034065", false},
		{"Provincia inválida (35)", "3510034065", false},
		{"Tercer dígito inválido (>=6)", "1770034065", false},
		{"Dígito verificador incorrecto", "1710034069", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ValidarCedula(tt.cedula)
			if res != tt.valida {
				t.Errorf("ValidarCedula(%s) = %v; esperaba %v", tt.cedula, res, tt.valida)
			}
		})
	}
}

func TestValidarPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valida   bool
	}{
		{"Válida con letras y números", "Password123", true},
		{"Válida mínima 8", "a1b2c3d4", true},
		{"Corta (<8)", "Pass1", false},
		{"Solo letras", "abcdefgh", false},
		{"Solo números", "12345678", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ValidarPassword(tt.password)
			if res != tt.valida {
				t.Errorf("ValidarPassword(%s) = %v; esperaba %v", tt.password, res, tt.valida)
			}
		})
	}
}

func TestEnmascararCedula(t *testing.T) {
	res := EnmascararCedula("1710034065")
	if res != "17******65" {
		t.Errorf("EnmascararCedula(1710034065) = %s; esperaba 17******65", res)
	}
}
