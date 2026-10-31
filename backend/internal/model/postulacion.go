package model

import "time"

// PostulacionHeader contiene los datos principales de la postulación.
type PostulacionHeader struct {
	ID                  int        `json:"id"`
	Codigo              string     `json:"codigo"`
	PostulanteID        int        `json:"postulante_id"`
	Cargo               *string    `json:"cargo"`
	RespaldoOrg         *bool      `json:"respaldo_org"`
	OrgNombre           *string    `json:"org_nombre"`
	OrgDocumento        *string    `json:"org_documento"`
	Estado              string     `json:"estado"`
	AceptaTerminos      bool       `json:"acepta_terminos"`
	FechaConsentimiento *time.Time `json:"fecha_consentimiento"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// Domicilio contiene la información de contacto y ubicación física.
type Domicilio struct {
	PostulacionID   int     `json:"postulacion_id,omitempty"`
	PaisID          int     `json:"pais_id"`
	ProvinciaID     *int    `json:"provincia_id,omitempty"`
	CantonID        *int    `json:"canton_id,omitempty"`
	ProvinciaTexto  *string `json:"provincia_texto,omitempty"`
	CiudadTexto     *string `json:"ciudad_texto,omitempty"`
	Celular         string  `json:"celular"`
	Email2          *string `json:"email2,omitempty"`
	CallePrincipal  string  `json:"calle_principal"`
	Numero          *string `json:"numero,omitempty"`
	CalleSecundaria *string `json:"calle_secundaria,omitempty"`
	Sector          *string `json:"sector,omitempty"`
	Referencia      *string `json:"referencia,omitempty"`
	Telefono        *string `json:"telefono,omitempty"`
}

// FamiliarIESS representa a un familiar que trabaja en IESS o BIESS.
type FamiliarIESS struct {
	ID             int     `json:"id,omitempty"`
	PostulacionID  int     `json:"postulacion_id,omitempty"`
	Nombres        string  `json:"nombres"`
	Parentesco     string  `json:"parentesco"`
	Institucion    string  `json:"institucion"`
	Area           *string `json:"area,omitempty"`
	Cargo          *string `json:"cargo,omitempty"`
}

// FormacionAcademica representa un título de tercer o cuarto nivel.
type FormacionAcademica struct {
	ID             int        `json:"id,omitempty"`
	PostulacionID  int        `json:"postulacion_id,omitempty"`
	Nivel          string     `json:"nivel"`
	Institucion    string     `json:"institucion"`
	Tipo           string     `json:"tipo"`
	Titulo         string     `json:"titulo"`
	Area           string     `json:"area"`
	SenescytNumero *string    `json:"senescyt_numero,omitempty"`
	SenescytFecha  *time.Time `json:"senescyt_fecha,omitempty"`
	Documento      *string    `json:"documento,omitempty"`
}

// ExperienciaLaboral representa experiencia general o de dirección.
type ExperienciaLaboral struct {
	ID             int        `json:"id,omitempty"`
	PostulacionID  int        `json:"postulacion_id,omitempty"`
	Tipo           string     `json:"tipo"`
	Cargo          string     `json:"cargo"`
	Institucion    string     `json:"institucion"`
	FechaInicio    time.Time  `json:"fecha_inicio"`
	FechaFin       *time.Time `json:"fecha_fin,omitempty"`
	Area           string     `json:"area"`
	Funciones      *string    `json:"funciones,omitempty"`
	Documento      *string    `json:"documento,omitempty"`
	DuracionTexto  string     `json:"duracion_texto,omitempty"` // Calculado (años, meses, días)
}

// Publicacion representa una publicación o trabajo de investigación académica.
type Publicacion struct {
	ID                  int       `json:"id,omitempty"`
	PostulacionID       int       `json:"postulacion_id,omitempty"`
	Titulo              string    `json:"titulo"`
	Tipo                string    `json:"tipo"`
	Medio               string    `json:"medio"`
	Fecha               time.Time `json:"fecha"`
	RelacionProfesional bool      `json:"relacion_profesional"`
	Documento           *string   `json:"documento,omitempty"`
}

// Capacitacion representa un curso, seminario o taller.
type Capacitacion struct {
	ID            int       `json:"id,omitempty"`
	PostulacionID int       `json:"postulacion_id,omitempty"`
	TipoEvento    string    `json:"tipo_evento"`
	NombreEvento  string    `json:"nombre_evento"`
	Institucion   string    `json:"institucion"`
	FechaInicio   time.Time `json:"fecha_inicio"`
	FechaFin      time.Time `json:"fecha_fin"`
	Horas         int       `json:"horas"`
	Documento     *string   `json:"documento,omitempty"`
}

// Inhabilidades contiene las 13 declaraciones juradas (EF-12 1.9).
type Inhabilidades struct {
	PostulacionID                   int        `json:"postulacion_id,omitempty"`
	GoceDerechosPoliticos           bool       `json:"goce_derechos_politicos"`
	InhabilitadoComercio            bool       `json:"inhabilitado_comercio"`
	MoraObligacionesEstado          bool       `json:"mora_obligaciones_estado"`
	VinculoSistemaFinanciero        bool       `json:"vinculo_sistema_financiero"`
	FuncionarioIessBiess            bool       `json:"funcionario_iess_biess"`
	InteresAseguradoras             bool       `json:"interes_aseguradoras"`
	RemovidoControl                 bool       `json:"removido_control"`
	RemovidoInstitucion             *string    `json:"removido_institucion,omitempty"`
	RemovidoCargo                   *string    `json:"removido_cargo,omitempty"`
	RemovidoFechaFin                *time.Time `json:"removido_fecha_fin,omitempty"`
	RemovidoMotivo                  *string    `json:"removido_motivo,omitempty"`
	ObligacionesSRI                 bool       `json:"obligaciones_sri"`
	CastigoObligacionesFinancieras  bool       `json:"castigo_obligaciones_financieras"`
	LitigioIessBiess                bool       `json:"litigio_iess_biess"`
	ProcesadoCorrupcion             bool       `json:"procesado_corrupcion"`
	ResponsabilidadContraloria      bool       `json:"responsabilidad_contraloria"`
	RegistradoUAFE                  bool       `json:"registrado_uafe"`
}

// PostulacionCompleta agrupa todas las secciones del formulario EF-12.
type PostulacionCompleta struct {
	PostulacionHeader
	Domicilio     *Domicilio           `json:"domicilio,omitempty"`
	Familiares    []FamiliarIESS       `json:"familiares,omitempty"`
	Formacion     []FormacionAcademica `json:"formacion,omitempty"`
	Experiencia   []ExperienciaLaboral `json:"experiencia,omitempty"`
	Publicaciones []Publicacion        `json:"publicaciones,omitempty"`
	Capacitaciones []Capacitacion      `json:"capacitaciones,omitempty"`
	Inhabilidades *Inhabilidades       `json:"inhabilidades,omitempty"`
}

// ValidarCargoRequest petición para validar aptitud para un cargo.
type ValidarCargoRequest struct {
	Cargo string `json:"cargo" binding:"required"`
}

// ValidarCargoResponse respuesta del chequeo de cargo.
type ValidarCargoResponse struct {
	Habilitado bool   `json:"habilitado"`
	Mensaje    string `json:"mensaje"`
}

// EnviarPostulacionRequest petición para enviar la postulación.
type EnviarPostulacionRequest struct {
	AceptaTerminos bool `json:"acepta_terminos"`
}
