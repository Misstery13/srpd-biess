package model

import "time"

// PostulantePerfil contiene los datos de solo lectura del Registro Civil (P6).
type PostulantePerfil struct {
	PostulanteID    int       `json:"postulante_id"`
	Email           string    `json:"email"`
	Cedula          string    `json:"cedula"`
	Nombres         string    `json:"nombres"`
	Apellidos       string    `json:"apellidos"`
	Genero          string    `json:"genero"`
	FechaNacimiento time.Time `json:"fecha_nacimiento"`
	Nacionalidad    string    `json:"nacionalidad"`
	EstadoCivil     string    `json:"estado_civil"`
	ConyugeNombres  *string   `json:"conyuge_nombres,omitempty"`
	ConyugeCedula   *string   `json:"conyuge_cedula,omitempty"`
	FechaExpedicion time.Time `json:"-"`
	CodigoDactilar  string    `json:"-"`
	EsAfiliado      bool      `json:"es_afiliado"`
	EsJubilado      bool      `json:"es_jubilado"`
	EsEmpleador     bool      `json:"es_empleador"`
}
