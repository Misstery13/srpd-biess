package cache

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// ==============================================================================
// Módulo Criptográfico para Protección de Caché en Redis
// Criterio: OWASP A02 (Cryptographic Failures) & LOPDP Art. 40
// Responsable: Gino Bermúdez (@ginobermudez76) — Issue #24
// ==============================================================================

var (
	ErrCiphertextInvalido = errors.New("formato de texto cifrado inválido")
	ErrAutenticacionFallo = errors.New("fallo de autenticación en descifrado AES-GCM (datos corruptos o manipulados)")
)

// DerivarClaveAES32 genera una clave simétrica de 256 bits (32 bytes) a partir del secreto del sistema.
func DerivarClaveAES32(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// CifrarAESGCM cifra y autentica un texto en claro usando AES-256 en modo GCM (Galois/Counter Mode).
// Cada invocación utiliza un nonce criptográfico de 12 bytes generado aleatoriamente.
// Devuelve una cadena en base64 con el formato: [nonce (12B)][ciphertext + tag (16B)].
func CifrarAESGCM(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error al inicializar bloque AES: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error al inicializar modo GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("error al generar nonce aleatorio: %w", err)
	}

	// El nonce se antepone al ciphertext para ser recuperado durante el descifrado
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DescifrarAESGCM extrae el nonce y descifra/autentica el texto cifrado en base64.
func DescifrarAESGCM(key []byte, ciphertextBase64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("error decodificando base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error al inicializar bloque AES: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error al inicializar modo GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrCiphertextInvalido
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrAutenticacionFallo
	}

	return string(plaintext), nil
}

// PseudonimizarClave genera un identificador ciego irreversible a partir de un valor sensible
// (p. ej. cédula de identidad) usando HMAC-SHA256 con el secreto del sistema.
// Esto evita que nombres de claves en Redis (`intentos:login:1710034065`) expongan cédulas en texto plano.
func PseudonimizarClave(secret, identificador string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(identificador))
	hash := mac.Sum(nil)
	// Truncado seguro a 16 bytes en hexadecimal (32 caracteres) para optimizar memoria en Redis
	return hex.EncodeToString(hash[:16])
}

