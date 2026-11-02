package cache

import (
	"strings"
	"testing"
)

func TestCrypto_CifradoYDescifradoAESGCM(t *testing.T) {
	secret := "mi-clave-super-secreta-de-prueba-biess-2026"
	key := DerivarClaveAES32(secret)

	cedula := "1710034065"

	// 1. Cifrado
	cifrado1, err := CifrarAESGCM(key, cedula)
	if err != nil {
		t.Fatalf("error al cifrar: %v", err)
	}
	if cifrado1 == "" || cifrado1 == cedula {
		t.Fatalf("el texto cifrado no puede ser vacío ni igual al original: %s", cifrado1)
	}

	// 2. Descifrado correcto
	descifrado, err := DescifrarAESGCM(key, cifrado1)
	if err != nil {
		t.Fatalf("error al descifrar: %v", err)
	}
	if descifrado != cedula {
		t.Errorf("se esperaba '%s', se obtuvo '%s'", cedula, descifrado)
	}

	// 3. Propiedad de No-Determinismo (Nonces aleatorios únicos):
	// Cifrar el mismo dato dos veces debe producir dos ciphertexts diferentes
	cifrado2, err := CifrarAESGCM(key, cedula)
	if err != nil {
		t.Fatalf("error en segundo cifrado: %v", err)
	}
	if cifrado1 == cifrado2 {
		t.Errorf("AES-GCM debe usar nonces aleatorios únicos por operación, ambos ciphertexts fueron iguales")
	}

	// 4. Integridad y Autenticación (Detección de Manipulación / Tampering):
	// Alterar un carácter del base64 debe disparar error de autenticación
	corrupto := cifrado1[:len(cifrado1)-4] + "AAAA"
	_, errCorrupto := DescifrarAESGCM(key, corrupto)
	if errCorrupto == nil {
		t.Errorf("se esperaba error de autenticación al intentar descifrar datos manipulados")
	}
}

func TestCrypto_PseudonimizarClave(t *testing.T) {
	secret := "jwt-secret-de-prueba"
	cedula := "1710034065"

	pseudo1 := PseudonimizarClave(secret, cedula)
	pseudo2 := PseudonimizarClave(secret, cedula)

	// Pseudonimización debe ser determinista para poder buscar en Redis
	if pseudo1 != pseudo2 {
		t.Errorf("la pseudonimización con la misma clave debe ser determinista")
	}

	// Longitud de 32 caracteres hexadecimales
	if len(pseudo1) != 32 {
		t.Errorf("se esperaban 32 caracteres hex, se obtuvieron %d", len(pseudo1))
	}

	// No debe contener la cédula original
	if strings.Contains(pseudo1, cedula) {
		t.Errorf("el pseudónimo no debe contener la cédula original en claro")
	}

	// Dos cédulas distintas deben producir pseudónimos distintos
	pseudoOtra := PseudonimizarClave(secret, "0912345675")
	if pseudo1 == pseudoOtra {
		t.Errorf("cédulas distintas produjeron el mismo pseudónimo (colisión inesperada)")
	}
}
