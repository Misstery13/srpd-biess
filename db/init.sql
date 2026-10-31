-- SRPD BIESS — DDL y datos semilla
-- Se ejecuta automáticamente al levantar docker-compose por primera vez.

BEGIN;

-- =========================================================
-- Catálogos geográficos (País -> Provincia -> Cantón)
-- =========================================================
CREATE TABLE paises (
    id     SERIAL PRIMARY KEY,
    nombre VARCHAR(80) NOT NULL UNIQUE
);

CREATE TABLE provincias (
    id      SERIAL PRIMARY KEY,
    pais_id INT NOT NULL REFERENCES paises(id),
    nombre  VARCHAR(80) NOT NULL,
    UNIQUE (pais_id, nombre)
);

CREATE TABLE cantones (
    id           SERIAL PRIMARY KEY,
    provincia_id INT NOT NULL REFERENCES provincias(id),
    nombre       VARCHAR(80) NOT NULL,
    UNIQUE (provincia_id, nombre)
);

-- =========================================================
-- Simulación de interoperabilidad (Registro Civil + Historial Laboral IESS)
-- =========================================================
CREATE TABLE registro_civil (
    cedula            CHAR(10) PRIMARY KEY,
    nombres           VARCHAR(100) NOT NULL,
    apellidos         VARCHAR(100) NOT NULL,
    genero            VARCHAR(20)  NOT NULL,
    fecha_nacimiento  DATE         NOT NULL,
    nacionalidad      VARCHAR(50)  NOT NULL,
    estado_civil      VARCHAR(30)  NOT NULL,
    conyuge_nombres   VARCHAR(200),
    conyuge_cedula    CHAR(10),
    fecha_expedicion  DATE         NOT NULL,
    codigo_dactilar   CHAR(10)     NOT NULL,
    es_afiliado       BOOLEAN NOT NULL DEFAULT FALSE,
    es_jubilado       BOOLEAN NOT NULL DEFAULT FALSE,
    es_empleador      BOOLEAN NOT NULL DEFAULT FALSE
);

-- =========================================================
-- Postulantes y postulaciones (EF-11 / EF-12)
-- =========================================================
CREATE TABLE postulantes (
    id            SERIAL PRIMARY KEY,
    cedula        CHAR(10) NOT NULL UNIQUE REFERENCES registro_civil(cedula),
    email         VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE postulaciones (
    id            SERIAL PRIMARY KEY,
    codigo        VARCHAR(20) NOT NULL UNIQUE,  -- código de numeración único
    postulante_id INT NOT NULL REFERENCES postulantes(id) ON DELETE CASCADE,
    cargo         VARCHAR(20) CHECK (cargo IN ('AFILIADOS', 'JUBILADOS', 'EMPLEADORES')),
    respaldo_org  BOOLEAN,
    org_nombre    VARCHAR(200),
    org_documento VARCHAR(255),                 -- ruta al PDF (máx 4 MB)
    estado        VARCHAR(20) NOT NULL DEFAULT 'BORRADOR'
                  CHECK (estado IN ('BORRADOR', 'ENVIADA', 'CERRADA')),
    acepta_terminos BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE domicilios (
    postulacion_id   INT PRIMARY KEY REFERENCES postulaciones(id) ON DELETE CASCADE,
    pais_id          INT NOT NULL REFERENCES paises(id),
    provincia_id     INT REFERENCES provincias(id),   -- si país = Ecuador
    canton_id        INT REFERENCES cantones(id),     -- si país = Ecuador
    provincia_texto  VARCHAR(100),                    -- si país != Ecuador
    ciudad_texto     VARCHAR(100),                    -- si país != Ecuador
    celular          VARCHAR(15) NOT NULL,
    email2           VARCHAR(150),
    calle_principal  VARCHAR(150) NOT NULL,
    numero           VARCHAR(20),
    calle_secundaria VARCHAR(150),
    sector           VARCHAR(100),
    referencia       VARCHAR(255),
    telefono         VARCHAR(15),
    CHECK (canton_id IS NOT NULL OR ciudad_texto IS NOT NULL)
);

CREATE TABLE familiares_iess (
    id             SERIAL PRIMARY KEY,
    postulacion_id INT NOT NULL REFERENCES postulaciones(id) ON DELETE CASCADE,
    nombres        VARCHAR(200) NOT NULL,
    parentesco     VARCHAR(30)  NOT NULL,
    institucion    VARCHAR(5)   NOT NULL CHECK (institucion IN ('IESS', 'BIESS')),
    area           VARCHAR(100),
    cargo          VARCHAR(100)
);

CREATE TABLE formacion_academica (
    id               SERIAL PRIMARY KEY,
    postulacion_id   INT NOT NULL REFERENCES postulaciones(id) ON DELETE CASCADE,
    nivel            VARCHAR(50)  NOT NULL,
    institucion      VARCHAR(150) NOT NULL,
    tipo             VARCHAR(10)  NOT NULL CHECK (tipo IN ('NACIONAL', 'EXTRANJERA')),
    titulo           VARCHAR(200) NOT NULL,
    area             VARCHAR(50)  NOT NULL,
    senescyt_numero  VARCHAR(50),
    senescyt_fecha   DATE,
    documento        VARCHAR(255)                     -- PDF máx 2 MB
);

CREATE TABLE experiencia_laboral (
    id             SERIAL PRIMARY KEY,
    postulacion_id INT NOT NULL REFERENCES postulaciones(id) ON DELETE CASCADE,
    tipo           VARCHAR(10)  NOT NULL CHECK (tipo IN ('GENERAL', 'DIRECCION')),
    cargo          VARCHAR(150) NOT NULL,
    institucion    VARCHAR(150) NOT NULL,
    fecha_inicio   DATE NOT NULL,
    fecha_fin      DATE,
    area           VARCHAR(50)  NOT NULL,
    funciones      VARCHAR(1000),
    documento      VARCHAR(255),
    CHECK (fecha_fin IS NULL OR fecha_fin >= fecha_inicio)
);

-- =========================================================
-- Semilla: países (lista del requerimiento EF-12 1.2.1)
-- =========================================================
INSERT INTO paises (nombre) VALUES
('Afganistán'),('Albania'),('Alemania'),('Andorra'),('Angola'),('Antigua y Barbuda'),
('Arabia Saudita'),('Argelia'),('Argentina'),('Armenia'),('Australia'),('Austria'),
('Azerbaiyán'),('Bahamas'),('Bangladés'),('Barbados'),('Baréin'),('Bélgica'),('Belice'),
('Benín'),('Bielorrusia'),('Birmania'),('Bolivia'),('Bosnia y Herzegovina'),('Botsuana'),
('Brasil'),('Brunéi'),('Bulgaria'),('Burkina Faso'),('Burundi'),('Bután'),('Cabo Verde'),
('Camboya'),('Camerún'),('Canadá'),('Catar'),('Chad'),('Chile'),('China'),('Chipre'),
('Ciudad del Vaticano'),('Colombia'),('Comoras'),('Corea del Norte'),('Corea del Sur'),
('Costa de Marfil'),('Costa Rica'),('Croacia'),('Cuba'),('Dinamarca'),('Dominica'),
('Ecuador'),('Egipto'),('El Salvador'),('Emiratos Árabes Unidos'),('Eritrea'),('Eslovaquia'),
('Eslovenia'),('España'),('Estados Unidos'),('Estonia'),('Etiopía'),('Filipinas'),
('Finlandia'),('Fiyi'),('Francia'),('Gabón'),('Gambia'),('Georgia'),('Ghana'),('Granada'),
('Grecia'),('Guatemala'),('Guyana'),('Guinea'),('Guinea-Bisáu'),('Guinea Ecuatorial'),
('Haití'),('Honduras'),('Hungría'),('India'),('Indonesia'),('Irak'),('Irán'),('Irlanda'),
('Islandia'),('Islas Marshall'),('Islas Salomón'),('Israel'),('Italia'),('Jamaica'),
('Japón'),('Jordania'),('Kazajistán'),('Kenia'),('Kirguistán'),('Kiribati'),('Kuwait'),
('Laos'),('Lesoto'),('Letonia'),('Líbano'),('Liberia'),('Libia'),('Liechtenstein'),
('Lituania'),('Luxemburgo'),('Macedonia del Norte'),('Madagascar'),('Malasia'),('Malaui'),
('Maldivas'),('Malí'),('Malta'),('Marruecos'),('Mauricio'),('Mauritania'),('México'),
('Micronesia'),('Moldavia'),('Mónaco'),('Mongolia'),('Montenegro'),('Mozambique'),
('Namibia'),('Nauru'),('Nepal'),('Nicaragua'),('Níger'),('Nigeria'),('Noruega'),
('Nueva Zelanda'),('Omán'),('Países Bajos'),('Pakistán'),('Palaos'),('Panamá'),
('Papúa Nueva Guinea'),('Paraguay'),('Perú'),('Polonia'),('Portugal'),('Reino Unido'),
('República Centroafricana'),('República Checa'),('República del Congo'),
('República Democrática del Congo'),('República Dominicana'),('Ruanda'),('Rumanía'),
('Rusia'),('Samoa'),('San Cristóbal y Nieves'),('San Marino'),('San Vicente y las Granadinas'),
('Santa Lucía'),('Santo Tomé y Príncipe'),('Senegal'),('Serbia'),('Seychelles'),
('Sierra Leona'),('Singapur'),('Siria'),('Somalia'),('Sri Lanka'),('Suazilandia'),
('Sudáfrica'),('Sudán'),('Sudán del Sur'),('Suecia'),('Suiza'),('Surinam'),('Tailandia'),
('Tanzania'),('Tayikistán'),('Timor Oriental'),('Togo'),('Tonga'),('Trinidad y Tobago'),
('Túnez'),('Turkmenistán'),('Turquía'),('Tuvalu'),('Ucrania'),('Uganda'),('Uruguay'),
('Uzbekistán'),('Vanuatu'),('Venezuela'),('Vietnam'),('Yemen'),('Yibuti'),('Zambia'),
('Zimbabue');

-- =========================================================
-- Semilla: provincias y cantones de Ecuador (EF-12 1.2.2 / 1.2.3)
-- =========================================================
INSERT INTO provincias (pais_id, nombre)
SELECT p.id, v.nombre
FROM paises p,
(VALUES ('Azuay'),('Bolívar'),('Cañar'),('Carchi'),('Chimborazo'),('Cotopaxi'),('El Oro'),
        ('Esmeraldas'),('Galápagos'),('Guayas'),('Imbabura'),('Loja'),('Los Ríos'),('Manabí'),
        ('Morona Santiago'),('Napo'),('Orellana'),('Pastaza'),('Pichincha'),('Santa Elena'),
        ('Santo Domingo de los Tsáchilas'),('Sucumbíos'),('Tungurahua'),('Zamora Chinchipe')
) AS v(nombre)
WHERE p.nombre = 'Ecuador';

INSERT INTO cantones (provincia_id, nombre)
SELECT pr.id, v.canton
FROM (VALUES
('Azuay','Cuenca'),('Azuay','Gualaceo'),('Azuay','Chordeleg'),('Azuay','Sígsig'),('Azuay','Girón'),
('Azuay','Santa Isabel'),('Azuay','Pucará'),('Azuay','Camilo Ponce Enríquez'),
('Bolívar','Guaranda'),('Bolívar','Chillanes'),('Bolívar','Chimbo'),('Bolívar','Echeandía'),
('Bolívar','San Miguel'),('Bolívar','Caluma'),('Bolívar','Las Naves'),
('Cañar','Azogues'),('Cañar','Biblián'),('Cañar','Cañar'),('Cañar','Déleg'),('Cañar','El Tambo'),
('Cañar','Suscal'),('Cañar','La Troncal'),
('Carchi','Tulcán'),('Carchi','Bolívar'),('Carchi','Espejo'),('Carchi','Mira'),('Carchi','Montúfar'),
('Carchi','San Pedro de Huaca'),
('Chimborazo','Riobamba'),('Chimborazo','Alausí'),('Chimborazo','Colta'),('Chimborazo','Chunchi'),
('Chimborazo','Guamote'),('Chimborazo','Guano'),('Chimborazo','Pallatanga'),('Chimborazo','Penipe'),
('Chimborazo','Cumandá'),
('Cotopaxi','Latacunga'),('Cotopaxi','La Maná'),('Cotopaxi','Pangua'),('Cotopaxi','Pujilí'),
('Cotopaxi','Salcedo'),('Cotopaxi','Saquisilí'),('Cotopaxi','Sigchos'),
('El Oro','Machala'),('El Oro','Arenillas'),('El Oro','Atahualpa'),('El Oro','Balsas'),('El Oro','Chilla'),
('El Oro','El Guabo'),('El Oro','Huaquillas'),('El Oro','Las Lajas'),('El Oro','Marcabelí'),
('El Oro','Pasaje'),('El Oro','Piñas'),('El Oro','Portovelo'),('El Oro','Santa Rosa'),('El Oro','Zaruma'),
('Esmeraldas','Esmeraldas'),('Esmeraldas','Atacames'),('Esmeraldas','Eloy Alfaro'),('Esmeraldas','Muisne'),
('Esmeraldas','Quinindé'),('Esmeraldas','Rioverde'),('Esmeraldas','San Lorenzo'),
('Galápagos','San Cristóbal'),('Galápagos','Isabela'),('Galápagos','Santa Cruz'),
('Guayas','Guayaquil'),('Guayas','Durán'),('Guayas','Milagro'),('Guayas','Samborondón'),('Guayas','Daule'),
('Guayas','Playas'),('Guayas','Balzar'),('Guayas','Colimes'),('Guayas','El Empalme'),('Guayas','El Triunfo'),
('Guayas','Isidro Ayora'),('Guayas','Lomas de Sargentillo'),('Guayas','Naranjal'),('Guayas','Naranjito'),
('Guayas','Nobol'),('Guayas','Pedro Carbo'),('Guayas','Salitre'),('Guayas','Simón Bolívar'),
('Guayas','Yaguachi'),
('Imbabura','Ibarra'),('Imbabura','Antonio Ante'),('Imbabura','Cotacachi'),('Imbabura','Otavalo'),
('Imbabura','Pimampiro'),('Imbabura','San Miguel de Urcuquí'),
('Loja','Loja'),('Loja','Calvas'),('Loja','Catamayo'),('Loja','Celica'),('Loja','Chaguarpamba'),
('Loja','Espíndola'),('Loja','Gonzanamá'),('Loja','Macará'),('Loja','Paltas'),('Loja','Puyango'),
('Loja','Saraguro'),('Loja','Sozoranga'),('Loja','Zapotillo'),
('Los Ríos','Babahoyo'),('Los Ríos','Baba'),('Los Ríos','Buena Fe'),('Los Ríos','Mocache'),
('Los Ríos','Montalvo'),('Los Ríos','Palenque'),('Los Ríos','Puebloviejo'),('Los Ríos','Quevedo'),
('Los Ríos','Urdaneta'),('Los Ríos','Ventanas'),('Los Ríos','Vinces'),('Los Ríos','Valencia'),
('Manabí','Portoviejo'),('Manabí','Manta'),('Manabí','Chone'),('Manabí','Bolívar'),('Manabí','El Carmen'),
('Manabí','Flavio Alfaro'),('Manabí','Jipijapa'),('Manabí','Junín'),('Manabí','Montecristi'),
('Manabí','Paján'),('Manabí','Pichincha'),('Manabí','Rocafuerte'),('Manabí','Santa Ana'),('Manabí','Sucre'),
('Manabí','Tosagua'),('Manabí','24 de Mayo'),('Manabí','Pedernales'),('Manabí','Olmedo'),
('Manabí','Puerto López'),('Manabí','Jama'),('Manabí','Jaramijó'),('Manabí','San Vicente'),
('Morona Santiago','Morona'),('Morona Santiago','Gualaquiza'),('Morona Santiago','Limón Indanza'),
('Morona Santiago','Palora'),('Morona Santiago','San Juan Bosco'),('Morona Santiago','Santiago'),
('Morona Santiago','Sucúa'),('Morona Santiago','Taisha'),('Morona Santiago','Tiwintza'),
('Napo','Tena'),('Napo','Archidona'),('Napo','Carlos Julio Arosemena Tola'),('Napo','El Chaco'),
('Napo','Quijos'),
('Orellana','Francisco de Orellana'),('Orellana','Aguarico'),('Orellana','La Joya de los Sachas'),
('Orellana','Loreto'),
('Pastaza','Pastaza'),('Pastaza','Arajuno'),('Pastaza','Mera'),('Pastaza','Santa Clara'),
('Pichincha','Quito'),('Pichincha','Cayambe'),('Pichincha','Mejía'),('Pichincha','Pedro Moncayo'),
('Pichincha','Pedro Vicente Maldonado'),('Pichincha','Puerto Quito'),('Pichincha','Rumiñahui'),
('Pichincha','San Miguel de los Bancos'),
('Santa Elena','Santa Elena'),('Santa Elena','La Libertad'),
('Santo Domingo de los Tsáchilas','La Concordia'),('Santo Domingo de los Tsáchilas','Santo Domingo'),
('Sucumbíos','Cascales'),('Sucumbíos','Cuyabeno'),('Sucumbíos','Gonzalo Pizarro'),
('Sucumbíos','Lago Agrio'),('Sucumbíos','Putumayo'),('Sucumbíos','Shushufindi'),('Sucumbíos','Sucumbíos'),
('Tungurahua','Ambato'),('Tungurahua','Baños de Agua Santa'),('Tungurahua','Cevallos'),
('Tungurahua','Mocha'),('Tungurahua','Patate'),('Tungurahua','Pelileo'),('Tungurahua','Píllaro'),
('Tungurahua','Quero'),('Tungurahua','Tisaleo'),
('Zamora Chinchipe','Centinela del Cóndor'),('Zamora Chinchipe','Chinchipe'),('Zamora Chinchipe','El Pangui'),
('Zamora Chinchipe','Nangaritza'),('Zamora Chinchipe','Palanda'),('Zamora Chinchipe','Paquisha'),
('Zamora Chinchipe','Yacuambi'),('Zamora Chinchipe','Yantzaza'),('Zamora Chinchipe','Zamora')
) AS v(provincia, canton)
JOIN provincias pr ON pr.nombre = v.provincia;

-- =========================================================
-- Semilla: personas simuladas del Registro Civil (datos ficticios para pruebas)
-- =========================================================
INSERT INTO registro_civil
(cedula, nombres, apellidos, genero, fecha_nacimiento, nacionalidad, estado_civil,
 conyuge_nombres, conyuge_cedula, fecha_expedicion, codigo_dactilar, es_afiliado, es_jubilado, es_empleador)
VALUES
('1710034065','JUAN CARLOS','PÉREZ LÓPEZ','MASCULINO','1980-05-12','ECUATORIANA','CASADO',
 'MARÍA ELENA TORRES','1712345678','2020-03-15','V4443V4444', TRUE,  FALSE, FALSE),
('0912345675','ANA LUCÍA','MORA VEGA','FEMENINO','1955-09-30','ECUATORIANA','SOLTERO',
 NULL, NULL,'2019-11-02','E1234I5678', FALSE, TRUE,  FALSE),
('0102030400','PEDRO','SALAZAR RUIZ','MASCULINO','1975-01-20','ECUATORIANA','DIVORCIADO',
 NULL, NULL,'2021-07-08','A1111B2222', FALSE, FALSE, TRUE),
('1754321980','CARLOS','GÓMEZ DÍAZ','MASCULINO','1985-02-14','COLOMBIANA','SOLTERO',
 NULL, NULL,'2022-01-10','C3333D4444', TRUE,  FALSE, FALSE);

COMMIT;
