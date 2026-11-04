package service

import (
	"fmt"
	"strconv"
	"unicode"
)

// ValidarCedula verifica el algoritmo módulo 10 para cédulas ecuatorianas (V1)
func ValidarCedula(cedula string) bool {
	if len(cedula) != 10 {
		return false
	}

	// Debe contener solo dígitos
	for _, c := range cedula {
		if !unicode.IsDigit(c) {
			return false
		}
	}

	// Provincia (dos primeros dígitos entre 01 y 24, o 30 para ecuatorianos en el exterior)
	provincia, err := strconv.Atoi(cedula[:2])
	if err != nil || ((provincia < 1 || provincia > 24) && provincia != 30) {
		return false
	}

	// Tercer dígito debe ser menor a 6 para personas naturales
	tercerDigito, err := strconv.Atoi(string(cedula[2]))
	if err != nil || tercerDigito >= 6 {
		return false
	}

	coeficientes := []int{2, 1, 2, 1, 2, 1, 2, 1, 2}
	suma := 0

	for i := 0; i < 9; i++ {
		digito, _ := strconv.Atoi(string(cedula[i]))
		prod := digito * coeficientes[i]
		if prod >= 10 {
			prod -= 9
		}
		suma += prod
	}

	digitoVerificadorCalculado := (10 - (suma % 10)) % 10
	digitoVerificadorReal, _ := strconv.Atoi(string(cedula[9]))

	return digitoVerificadorCalculado == digitoVerificadorReal
}

// ValidarPassword verifica longitud >= 8, al menos 1 letra y al menos 1 número (S3)
func ValidarPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	tieneLetra := false
	tieneNumero := false

	for _, c := range password {
		if unicode.IsLetter(c) {
			tieneLetra = true
		} else if unicode.IsDigit(c) {
			tieneNumero = true
		}
	}

	return tieneLetra && tieneNumero
}

// EnmascararCedula enmascara la cédula para registros de log seguros (O8) -> 17******65
func EnmascararCedula(cedula string) string {
	if len(cedula) < 4 {
		return "****"
	}
	return fmt.Sprintf("%s******%s", cedula[:2], cedula[len(cedula)-2:])
}
