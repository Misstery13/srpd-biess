package model

import "time"

// Periodo representa la respuesta pública del periodo de postulación (T2).
type Periodo struct {
	Abierto bool      `json:"abierto"`
	Inicio  time.Time `json:"inicio"`
	Fin     time.Time `json:"fin"`
}
