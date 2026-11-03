// V1: cédula ecuatoriana = provincia 01–24 (o 30), tercer dígito < 6 y dígito verificador módulo 10
export function cedulaValida(cedula) {
  if (!/^\d{10}$/.test(cedula)) return false
  const provincia = Number(cedula.slice(0, 2))
  if (!((provincia >= 1 && provincia <= 24) || provincia === 30) || Number(cedula[2]) >= 6) return false
  let suma = 0
  for (let i = 0; i < 9; i++) {
    let n = Number(cedula[i]) * (i % 2 === 0 ? 2 : 1)
    if (n > 9) n -= 9
    suma += n
  }
  return (10 - (suma % 10)) % 10 === Number(cedula[9])
}

// S3: mínimo 8 caracteres, al menos 1 letra y 1 número
export const passwordValida = (p) => p.length >= 8 && /[A-Za-z]/.test(p) && /\d/.test(p)

// V3: código dactilar = 10 caracteres alfanuméricos
export const codigoDactilarValido = (c) => /^[A-Z0-9]{10}$/.test(c)

export function emailValido(email) {
  const partes = email.split('@')
  if (partes.length !== 2 || /\s/.test(email)) return false
  const [usuario, dominio] = partes
  return usuario.length > 0 && dominio.includes('.') && !dominio.startsWith('.') && !dominio.endsWith('.')
}

// Mientras se escribe: deja solo dígitos e inserta las barras → DD/MM/AAAA
export function formatearFecha(texto) {
  const d = texto.replace(/\D/g, '').slice(0, 8)
  return [d.slice(0, 2), d.slice(2, 4), d.slice(4)].filter(Boolean).join('/')
}

// V2: DD/MM/AAAA (UI) → YYYY-MM-DD (API); null si la fecha no existe o es futura
export function fechaAIso(texto) {
  const m = /^(\d{2})\/(\d{2})\/(\d{4})$/.exec(texto)
  if (!m) return null
  const [, dd, mm, aaaa] = m
  const fecha = new Date(Date.UTC(aaaa, mm - 1, dd))
  if (fecha.getUTCDate() !== Number(dd) || fecha.getUTCMonth() !== mm - 1 || fecha > new Date()) return null
  return `${aaaa}-${mm}-${dd}`
}

export const IDENTIDAD_VACIA = { cedula: '', metodo: 'fecha', fecha: '', codigo: '' }

// Valida los campos de identidad (registro y recuperación) → { errores, datos } con el body del contrato
export function validarIdentidad({ cedula, metodo, fecha, codigo }) {
  const errores = {}
  if (!cedulaValida(cedula)) errores.cedula = 'Ingrese una cédula ecuatoriana válida.'
  const datos = { cedula }
  if (metodo === 'fecha') {
    datos.fecha_expedicion = fechaAIso(fecha)
    if (!datos.fecha_expedicion) errores.fecha = 'Ingrese una fecha válida con el formato DD/MM/AAAA.'
  } else {
    datos.codigo_dactilar = codigo
    if (!codigoDactilarValido(codigo)) errores.codigo = 'El código dactilar tiene 10 caracteres (letras y números).'
  }
  return { errores, datos }
}

export const validarPassword = (password, confirmar) => {
  const errores = {}
  if (!passwordValida(password)) errores.password = 'Mínimo 8 caracteres, con al menos una letra y un número.'
  if (password !== confirmar) errores.confirmar = 'Las contraseñas no coinciden.'
  return errores
}

// ===== Formatos de los campos de EF-12 (mismas reglas en backend/internal/service/campos.go) =====
// Nombres de personas y lugares: solo letras, espacios, apóstrofo, punto y guion; al menos 2 letras
export const esNombre = (t) => /^[\p{L}\p{M}' .-]+$/u.test(t) && (t.match(/\p{L}/gu) ?? []).length >= 2
// Texto general (instituciones, títulos, cargos, calles…): letras, números y puntuación común; debe tener letras
export const esTexto = (t) => /^[\p{L}\p{M}\p{N}\s.,;:()'"&/#°º+¿?¡!@_–-]+$/u.test(t) && /\p{L}/u.test(t)
// Número de casa: letras, números, espacio, / . - y al menos un dígito (o S/N)
export const esNumeroCasa = (t) => /^s\/?n$/i.test(t) || (/^[\p{L}\p{N} /.-]+$/u.test(t) && /\d/.test(t))
// Registro SENESCYT: letras, números y guiones, con al menos un dígito (p. ej. 1005-2018-1234567)
export const esSenescyt = (t) => /^[A-Za-z0-9-]+$/.test(t) && /\d/.test(t)
// Ecuador: celular 09XXXXXXXX y fijo 0[2-7]XXXXXXX; otro país: 7 a 15 dígitos con + opcional
export const esCelular = (t, ecuador) => (ecuador ? /^09\d{8}$/ : /^\+?\d{7,15}$/).test(t)
export const esTelefonoFijo = (t, ecuador) => (ecuador ? /^0[2-7]\d{7}$/ : /^\+?\d{7,15}$/).test(t)
// Fechas de <input type="date"> (YYYY-MM-DD): entre 1900 y hoy
export const fechaEnRango = (f, hoy) => f >= '1900-01-01' && f <= hoy

// Filtros mientras se escribe
export const sinDigitos = (t) => t.replace(/\d/g, '')
export const soloTelefono = (t) => t.replace(/[^\d+]/g, '').replace(/(?!^)\+/g, '')
export const soloDigitos = (t) => t.replace(/\D/g, '')

const MENSAJES = {
  nombre: 'Use solo letras (sin números ni símbolos).',
  texto: 'Debe contener letras; no use solo números ni símbolos como < > { } $ % * =.',
}
const VALIDADORES = { nombre: esNombre, texto: esTexto }

// Agrega a `e` el error de un campo de texto: vacío (si es obligatorio) o con formato inválido
export function revisarTexto(e, campo, valor, { vacio, formato = 'texto' } = {}) {
  const v = (valor ?? '').trim()
  if (!v) {
    if (vacio) e[campo] = vacio
  } else if (!VALIDADORES[formato](v)) e[campo] = MENSAJES[formato]
}
