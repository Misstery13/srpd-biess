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
