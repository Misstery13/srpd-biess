package model

type Pais struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type Provincia struct {
	ID     int    `json:"id"`
	PaisID int    `json:"pais_id,omitempty"`
	Nombre string `json:"nombre"`
}

type Canton struct {
	ID          int    `json:"id"`
	ProvinciaID int    `json:"provincia_id,omitempty"`
	Nombre      string `json:"nombre"`
}
