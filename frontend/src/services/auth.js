import api from './api.js'

// Mensajes exactos de EF-11 (se muestran tal cual, sin depender del texto del backend)
export const MENSAJES = {
  identidadIncorrecta: 'la información ingresada es incorrecta',
  registroBloqueado: 'registro fallido, inténtelo nuevamente en 10 minutos',
  credencialesIncorrectas: 'sus credenciales de ingreso son incorrectas',
  loginBloqueado: 'demasiados intentos fallidos, inténtelo nuevamente en 10 minutos',
}

// 401 y 429 muestran el mensaje exacto del documento; cualquier otro error, el del servidor
export function mensajeError(err, msg401, msg429 = err.message) {
  if (err.status === 401) return msg401
  if (err.status === 429) return msg429
  return err.message
}

const MOCK = import.meta.env.VITE_MOCK === 'true'

export async function verificar(datos) {
  if (MOCK) return mock.verificar(datos)
  return (await api.post('/auth/verificar', datos)).data
}

export async function registrar(datos) {
  if (MOCK) return mock.registrar(datos)
  await api.post('/auth/registro', datos)
}

export async function login(datos) {
  if (MOCK) return mock.login(datos)
  return (await api.post('/auth/login', datos)).data
}

export async function recuperar(datos) {
  if (MOCK) return mock.verificar(datos)
  await api.post('/auth/recuperar', datos)
}

export async function restablecer(datos) {
  if (MOCK) return mock.restablecer(datos)
  await api.post('/auth/restablecer', datos)
}

// ---- Simulación (VITE_MOCK=true): mismas personas ficticias que db/init.sql ----
const registroCivil = {
  1710034065: { nombres: 'JUAN CARLOS', fecha_expedicion: '2020-03-15', codigo_dactilar: 'V4443V4444' },
  '0912345675': { nombres: 'ANA LUCÍA', fecha_expedicion: '2019-11-02', codigo_dactilar: 'E1234I5678' },
  1754321980: { nombres: 'CARLOS', fecha_expedicion: '2022-01-10', codigo_dactilar: 'C3333D4444' },
}
const fallos = {}
const cuentas = {}
const falla = (status, error) => Promise.reject(Object.assign(new Error(error), { status }))
const demora = (data) => new Promise((r) => setTimeout(() => r(data), 300))

// S4/S5: 3 intentos fallidos → bloqueo de 10 min; un acierto reinicia el contador
function intento(clave, correcto, msg401, msg429) {
  const f = fallos[clave]
  if (f?.hasta > Date.now()) return falla(429, msg429)
  if (correcto) {
    delete fallos[clave]
    return null
  }
  const n = (f?.hasta ? 0 : f?.n ?? 0) + 1
  fallos[clave] = n >= 3 ? { n, hasta: Date.now() + 600_000 } : { n }
  return falla(n >= 3 ? 429 : 401, n >= 3 ? msg429 : msg401)
}

const mock = {
  verificar({ cedula, fecha_expedicion, codigo_dactilar }) {
    const p = registroCivil[cedula]
    const ok = p && (fecha_expedicion ? p.fecha_expedicion === fecha_expedicion : p.codigo_dactilar === codigo_dactilar)
    return intento(`registro:${cedula}`, ok, MENSAJES.identidadIncorrecta, MENSAJES.registroBloqueado) ?? demora({ nombres: p.nombres })
  },
  registrar({ cedula, password }) {
    if (cuentas[cedula]) return falla(409, 'Ya existe una cuenta registrada con esta cédula.')
    cuentas[cedula] = password
    return demora()
  },
  login({ cedula, password }) {
    const ok = cuentas[cedula] !== undefined && cuentas[cedula] === password
    return intento(`login:${cedula}`, ok, MENSAJES.credencialesIncorrectas, MENSAJES.loginBloqueado) ?? demora({ token: `mock.${cedula}` })
  },
  restablecer({ token }) {
    return token === 'mock' ? demora() : falla(400, 'El enlace de recuperación no es válido o ya expiró.')
  },
}
