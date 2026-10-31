package model

import "time"

type RegistroCivil struct {
	Cedula           string    `json:"cedula"`
	Nombres          string    `json:"nombres"`
	Apellidos        string    `json:"apellidos"`
	Genero           string    `json:"genero"`
	FechaNacimiento  time.Time `json:"fecha_nacimiento"`
	Nacionalidad     string    `json:"nacionalidad"`
	EstadoCivil      string    `json:"estado_civil"`
	ConyugeNombres   *string   `json:"conyuge_nombres,omitempty"`
	ConyugeCedula    *string   `json:"conyuge_cedula,omitempty"`
	FechaExpedicion  time.Time `json:"fecha_expedicion"`
	CodigoDactilar   string    `json:"codigo_dactilar"`
	EsAfiliado       bool      `json:"es_afiliado"`
	EsJubilado       bool      `json:"es_jubilado"`
	EsEmpleador      bool      `json:"es_empleador"`
}

type Postulante struct {
	ID           int        `json:"id"`
	Cedula       string     `json:"cedula"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	RolID        int        `json:"rol_id"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type Rol struct {
	ID        int        `json:"id"`
	Codigo    string     `json:"codigo"`
	Nombre    string     `json:"nombre"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Permiso struct {
	ID        int        `json:"id"`
	RolID     int        `json:"rol_id"`
	Metodo    string     `json:"metodo"`
	Ruta      string     `json:"ruta"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
