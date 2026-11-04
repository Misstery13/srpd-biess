import api, { getToken } from './api.js'

// Contrato: docs/plan/fase-0-acuerdos.md + modelos de backend/internal/model/postulacion.go
const MOCK = import.meta.env.VITE_MOCK === 'true'

export async function getPerfil() {
  if (MOCK) return mock.perfil()
  return (await api.get('/postulante/me')).data
}

export async function getPeriodo() {
  if (MOCK) return mock.periodo()
  return (await api.get('/periodo')).data
}

// null si todavía no existe (404)
export async function getPostulacion() {
  if (MOCK) return mock.obtener()
  try {
    return (await api.get('/postulacion')).data
  } catch (err) {
    if (err.status === 404) return null
    throw err
  }
}

export async function guardarPostulacion(postulacion) {
  if (MOCK) return mock.guardar(postulacion)
  return (await api.put('/postulacion', postulacion)).data // { codigo }
}

export async function validarCargo(cargo) {
  if (MOCK) return mock.validarCargo(cargo)
  return (await api.post('/postulacion/validar-cargo', { cargo })).data // { habilitado, mensaje }
}

export async function subirDocumento(archivo, tipo) {
  if (MOCK) return mock.subir()
  const datos = new FormData()
  datos.append('archivo', archivo)
  datos.append('tipo', tipo)
  return (await api.post('/postulacion/documentos', datos)).data.ruta
}

export async function enviarPostulacion() {
  if (MOCK) return mock.enviar()
  return (await api.post('/postulacion/enviar', { acepta_terminos: true })).data // { codigo }
}

export async function eliminarCuenta() {
  if (MOCK) return mock.eliminar()
  await api.delete('/postulante/me')
}

// ---- Simulación (VITE_MOCK=true): mismas personas ficticias que db/init.sql ----
const personas = {
  1710034065: { nombres: 'JUAN CARLOS', apellidos: 'PÉREZ LÓPEZ', genero: 'MASCULINO', fecha_nacimiento: '1980-05-12T00:00:00Z', nacionalidad: 'ECUATORIANA', estado_civil: 'CASADO', conyuge_nombres: 'MARÍA ELENA TORRES', conyuge_cedula: '1712345678', es_afiliado: true, es_jubilado: false, es_empleador: false },
  '0912345675': { nombres: 'ANA LUCÍA', apellidos: 'MORA VEGA', genero: 'FEMENINO', fecha_nacimiento: '1955-09-30T00:00:00Z', nacionalidad: 'ECUATORIANA', estado_civil: 'SOLTERO', es_afiliado: false, es_jubilado: true, es_empleador: false },
  1754321980: { nombres: 'CARLOS', apellidos: 'GÓMEZ DÍAZ', genero: 'MASCULINO', fecha_nacimiento: '1985-02-14T00:00:00Z', nacionalidad: 'COLOMBIANA', estado_civil: 'SOLTERO', es_afiliado: true, es_jubilado: false, es_empleador: false },
}
let guardada = null
const falla = (status, error) => Promise.reject(Object.assign(new Error(error), { status }))
const demora = (data) => new Promise((r) => setTimeout(() => r(data), 200))
const cedulaSesion = () => getToken()?.replace('mock.', '')
const leer = (clave) => {
  try {
    return sessionStorage.getItem(clave)
  } catch {
    return null
  }
}

const mock = {
  perfil() {
    const cedula = cedulaSesion()
    const p = personas[cedula]
    if (!p) return falla(404, 'Postulante no encontrado.')
    if (p.nacionalidad !== 'ECUATORIANA') return falla(403, 'Solo ciudadanos con nacionalidad ecuatoriana pueden postular.')
    return demora({ ...p, cedula, email: `${cedula}@example.com` })
  },
  periodo() {
    // sessionStorage "mock:periodo" = "cerrado" simula el período fuera de fecha (T2)
    const abierto = leer('mock:periodo') !== 'cerrado'
    return demora({ abierto, inicio: '2026-10-01T05:00:00Z', fin: '2026-12-01T04:59:00Z' })
  },
  obtener: () => demora(guardada && structuredClone(guardada)),
  guardar(p) {
    if (guardada?.estado === 'ENVIADA') return falla(403, 'Una postulación ya enviada no puede ser modificada (P3).')
    guardada = { ...structuredClone(p), codigo: guardada?.codigo ?? 'BIESS-2026-000001', estado: 'BORRADOR' }
    return demora({ codigo: guardada.codigo })
  },
  validarCargo(codigo) {
    const p = personas[cedulaSesion()]
    const habilitado = { AFILIADOS: p.es_afiliado, JUBILADOS: p.es_jubilado, EMPLEADORES: p.es_empleador }[codigo]
    return demora({ habilitado, mensaje: '' })
  },
  subir: () => demora(`uploads/1/${crypto.randomUUID()}.pdf`),
  enviar() {
    if (!guardada) return falla(404, 'No existe una postulación activa para enviar.')
    if (guardada.estado === 'ENVIADA') return falla(409, 'La postulación ya ha sido enviada anteriormente.')
    guardada.estado = 'ENVIADA'
    guardada.acepta_terminos = true
    return demora({ codigo: guardada.codigo })
  },
  eliminar() {
    guardada = null
    return demora()
  },
}
