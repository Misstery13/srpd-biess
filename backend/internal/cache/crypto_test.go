package cache

import (
	"strings"
	"testing"
)

func TestCrypto_CifradoYDescifradoAESGCM(t *testing.T) {
	key := DerivarClaveAES32("clave-prueba-para-aes-gcm-biess")
	cedula := "1710034065"

	t.Run("CifradoYDescifradoBasico", func(t *testing.T) {
		cifrado, err := CifrarAESGCM(key, cedula)
		if err != nil {
			t.Fatalf("error al cifrar dato: %v", err)
		}
		if cifrado == "" || cifrado == cedula {
			t.Fatalf("el texto cifrado no puede ser vacío ni igual al texto en claro: %s", cifrado)
		}

		descifrado, err := DescifrarAESGCM(key, cifrado)
		if err != nil {
			t.Fatalf("error al descifrar: %v", err)
		}
		if descifrado != cedula {
			t.Fatalf("descifrado incorrecto: esperado '%s', obtenido '%s'", cedula, descifrado)
		}
	})

	t.Run("NoDeterminismoConNoncesUnicos", func(t *testing.T) {
		c1, _ := CifrarAESGCM(key, cedula)
		c2, _ := CifrarAESGCM(key, cedula)
		if c1 == c2 {
			t.Errorf("AES-GCM debe usar nonces aleatorios únicos por operación")
		}
	})

	t.Run("DeteccionDeManipulacion", func(t *testing.T) {
		cifrado, _ := CifrarAESGCM(key, cedula)
		corrupto := cifrado[:len(cifrado)-4] + "AAAA"
		if _, err := DescifrarAESGCM(key, corrupto); err == nil {
			t.Errorf("se esperaba error de autenticación ante datos alterados")
		}
	})
}

func TestCrypto_PseudonimizarClave(t *testing.T) {
	secret := "jwt-secret-de-prueba"
	cedula := "1710034065"

	t.Run("Determinismo", func(t *testing.T) {
		p1 := PseudonimizarClave(secret, cedula)
		p2 := PseudonimizarClave(secret, cedula)
		if p1 != p2 || len(p1) != 32 {
			t.Errorf("la pseudonimización debe ser determinista y de 32 hex chars")
		}
	})

	t.Run("PrivacidadYNoColision", func(t *testing.T) {
		p1 := PseudonimizarClave(secret, cedula)
		if strings.Contains(p1, cedula) {
			t.Errorf("el pseudónimo no debe exponer la cédula")
		}

		pOtros := PseudonimizarClave(secret, "0912345675")
		if p1 == pOtros {
			t.Errorf("cédulas distintas no deben colisionar")
		}
	})
}
