// Listas, textos y reglas de EF-12 tal como aparecen en docs/requerimiento.pdf

// 1.4: cargos. `codigo` es el valor que guarda la BD (CHECK en postulaciones.cargo)
export const CARGOS = [
  {
    codigo: 'AFILIADOS',
    nombre: 'Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS',
    habilitado: 'Postulante afiliado al IESS',
    noHabilitado: 'Postulante No Afiliado al IESS',
    cierre:
      'Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS',
    respaldo:
      '¿Se encuentra respaldadas por una organización de trabajadores de representación nacional legalmente constituida ante autoridad competente?',
  },
  {
    codigo: 'JUBILADOS',
    nombre: 'Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS',
    habilitado: 'Postulante acredita jubilación',
    noHabilitado: 'Postulante no acredita jubilación',
    cierre:
      'Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación',
    respaldo:
      '¿Se encuentra respaldadas por una organización de jubilados de representación nacional legalmente constituida ante autoridad competente?',
  },
  {
    codigo: 'EMPLEADORES',
    nombre: 'Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS',
    habilitado: 'Postulante es empleador',
    noHabilitado: 'Postulante no es empleador',
    cierre:
      'Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador',
    respaldo:
      '¿Se encuentra respaldadas por una federación de cámara de representación nacional legalmente constituida ante autoridad competente?',
  },
]

export const cargoPorCodigo = (codigo) => CARGOS.find((c) => c.codigo === codigo)

// 1.3.1.2 y 1.3.1.3
export const PARENTESCOS = [
  'conviviente', 'cónyuge', 'cuñado/a', 'hermano/a', 'hijo político', 'hijo/a', 'madre', 'nieto/a',
  'nuera', 'padre', 'primo hermano/a', 'sobrino/a', 'suegro/a', 'tío/a', 'yerno',
]
export const INSTITUCIONES = ['IESS', 'BIESS']

// 1.5.1, 1.5.1.2 y 1.5.1.4 / 1.6.6 (misma lista de áreas)
export const NIVELES = [
  'Tercer Nivel de Grado',
  'Cuarto Nivel - Especialidad Académica',
  'Cuarto Nivel - Maestría Académica',
  'Cuarto Nivel PhD - Doctorado',
]
export const AREAS = ['Finanzas', 'Mercado de Capitales', 'Administración de Empresas', 'Economía', 'Derecho', 'Materias afines']

// Opciones con código de BD y texto visible
export const TIPOS_FORMACION = [['NACIONAL', 'Nacional'], ['EXTRANJERA', 'Extranjera']]
export const TIPOS_EXPERIENCIA = [['GENERAL', 'General'], ['DIRECCION', 'Dirección']]
export const TIPOS_PUBLICACION = [['PUBLICACION', 'Publicación'], ['INVESTIGACION', 'Trabajo de investigación']]
export const TIPOS_EVENTO = [['CURSO', 'Curso'], ['SEMINARIO', 'Seminario'], ['TALLER', 'Taller']]

// 1.9: las 13 declaraciones (clave = columna de la tabla inhabilidades)
export const INHABILIDADES = [
  ['goce_derechos_politicos', '¿Es Ecuatoriano/a y está en pleno goce de los derechos de participación política?'],
  ['inhabilitado_comercio', '¿Se encuentra inhabilitado para ejercer el comercio?'],
  ['mora_obligaciones_estado', '¿Está en mora, directa o indirectamente, en el pago de sus obligaciones en cualquiera de las instituciones del Estado o de las instituciones bajo control de la Superintendencia de Bancos, incluidas las off-shore?'],
  ['vinculo_sistema_financiero', '¿Mantiene algún vínculo, por propiedad o administración, con las instituciones del sistema financiero y del sistema de seguros privados, con las bolsas de valores del país, Casas de Valores y Administradoras de Fondos y Fideicomisos, de acuerdo con las normas que la Junta de Política y Regulación Monetaria y Financiera emita para el efecto?'],
  ['funcionario_iess_biess', '¿Es funcionario o empleado del IESS - BIESS?'],
  ['interes_aseguradoras', '¿Mantiene interés propio o representa a terceros en la propiedad, la dirección o la gestión de las compañías aseguradoras u otras personas que integran el sistema nacional de seguridad social?'],
  ['removido_control', '¿En los últimos 5 años, ha sido removido por algún Organismo de Control?'],
  ['obligaciones_sri', '¿Mantiene obligaciones pendientes con el SRI?'],
  ['castigo_obligaciones_financieras', '¿Ha incurrido en los últimos 5 años en castigo de sus obligaciones por parte de cualquier institución financiera?'],
  ['litigio_iess_biess', '¿Actualmente se encuentra litigando en contra del IESS- BIESS?'],
  ['procesado_corrupcion', '¿Se encuentra procesado, enjuiciado o condenado por algún delito relacionado con corrupción o crimen organizado?'],
  ['responsabilidad_contraloria', '¿Mantiene en firme algún tipo de responsabilidades por la Contraloría General del Estado?'],
  ['registrado_uafe', '¿Se encuentra registrado en la base de datos de la Unidad de Análisis Financiero y Económico?'],
]

// 1.10.1: los 5 textos de consentimiento (cada uno con su casilla, sin marcar por defecto, L1)
export const CONSENTIMIENTOS = [
  'Autorizo de manera libre y voluntaria, emitiendo consentimiento explícito al Instituto Ecuatoriano de Seguridad Social, para el uso de mis datos personales dentro del concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores, de conformidad con lo prescrito en el artículo 7 de la Ley Orgánica de Protección de Datos Personales.',
  'Declaro que la información del presente formulario y documentación adjunta corresponde a la verdad y autorizo al Instituto Ecuatoriano de Seguridad Social para que se compruebe ante cualquier organismo público o privado su veracidad.',
  'Acepto expresamente cumplir con las normas aplicables al concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores; así como las resoluciones y disposiciones impartidas por el Instituto Ecuatoriano de Seguridad Social. En caso de incumplimiento de requisitos o encontrarme inmerso/a en las prohibiciones o inhabilidades acepto mi descalificación del proceso.',
  'Acepto que mi nombre como postulante al concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores; sea publicado en la página web del IESS - BIESS.',
  'El presente formulario, debe ser impreso y con firma manuscrita, entregado junto con los requisitos solicitados, en el lugar, fecha y horario establecido por el Instituto Ecuatoriano de Seguridad Social.',
]

// A3: límites de los PDF
export const MB = 1024 * 1024
export const LIMITE_ORGANIZACION = 4 * MB
export const LIMITE_GENERAL = 2 * MB

// 1.1.7.1: se muestran los datos de la pareja si es casado/a o unión de hecho
export const tienePareja = (estadoCivil = '') => /^CASAD|UNI[OÓ]N DE HECHO/i.test(estadoCivil)

// La API (Go time.Time) usa RFC 3339; <input type="date"> usa YYYY-MM-DD
export const aFechaApi = (d) => (d ? `${d}T00:00:00Z` : null)
export const deFechaApi = (d) => (d ? d.slice(0, 10) : '')

const fechaLocal = (iso) => {
  const [a, m, d] = iso.split('-').map(Number)
  return new Date(Date.UTC(a, m - 1, d))
}

// 1.6: duración en años, meses y días entre dos fechas YYYY-MM-DD (sin fin = hasta hoy). No se guarda (V7).
export function calcularDuracion(inicio, fin, hoy = new Date()) {
  if (!inicio) return null
  const desde = fechaLocal(inicio)
  const hasta = fin ? fechaLocal(fin) : new Date(Date.UTC(hoy.getFullYear(), hoy.getMonth(), hoy.getDate()))
  if (hasta < desde) return null
  // Meses completos: el 31/01 + 1 mes es el 29/02 (se ajusta al último día del mes)
  const sumarMeses = (n) => {
    const a = desde.getUTCFullYear()
    const m = desde.getUTCMonth() + n
    const ultimo = new Date(Date.UTC(a, m + 1, 0)).getUTCDate()
    return new Date(Date.UTC(a, m, Math.min(desde.getUTCDate(), ultimo)))
  }
  let total = (hasta.getUTCFullYear() - desde.getUTCFullYear()) * 12 + hasta.getUTCMonth() - desde.getUTCMonth()
  if (sumarMeses(total) > hasta) total--
  const dias = Math.round((hasta - sumarMeses(total)) / 86_400_000)
  return { anios: Math.floor(total / 12), meses: total % 12, dias }
}

// Suma de duraciones normalizada (30 días = 1 mes es aproximado; solo para el total informativo)
export function sumarDuraciones(lista) {
  let { anios, meses, dias } = lista.filter(Boolean).reduce((t, d) => ({ anios: t.anios + d.anios, meses: t.meses + d.meses, dias: t.dias + d.dias }), { anios: 0, meses: 0, dias: 0 })
  meses += Math.floor(dias / 30)
  dias %= 30
  anios += Math.floor(meses / 12)
  meses %= 12
  return { anios, meses, dias }
}

const plural = (n, uno, varios) => `${n} ${n === 1 ? uno : varios}`
export const textoDuracion = (d) =>
  d ? `${plural(d.anios, 'año', 'años')}, ${plural(d.meses, 'mes', 'meses')} y ${plural(d.dias, 'día', 'días')}` : '—'
