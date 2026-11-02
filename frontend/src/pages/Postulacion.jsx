import { useCallback, useEffect, useState } from 'react'
import { Alerta } from '../components/Formulario.jsx'
import { useAuth } from '../context/auth.js'
import { getPeriodo, getPerfil, getPostulacion, guardarPostulacion } from '../services/postulacion.js'
import Capacitaciones from './postulacion/Capacitaciones.jsx'
import Cargo from './postulacion/Cargo.jsx'
import { fechaUI } from './postulacion/comun.js'
import Domicilio from './postulacion/Domicilio.jsx'
import EliminarCuenta from './postulacion/EliminarCuenta.jsx'
import Envio from './postulacion/Envio.jsx'
import Experiencia from './postulacion/Experiencia.jsx'
import Familiares from './postulacion/Familiares.jsx'
import Formacion from './postulacion/Formacion.jsx'
import Impresion from './postulacion/Impresion.jsx'
import InfoPersonal from './postulacion/InfoPersonal.jsx'
import Inhabilidades from './postulacion/Inhabilidades.jsx'
import Publicaciones from './postulacion/Publicaciones.jsx'

// Secciones de EF-12 en el orden del documento; `listo` marca la pestaña cuando ya se guardó algo
const SECCIONES = [
  { id: 'personal', titulo: 'Información personal', Componente: InfoPersonal, listo: () => true },
  { id: 'domicilio', titulo: 'Domicilio y contacto', Componente: Domicilio, listo: (p) => !!p?.domicilio },
  { id: 'adicional', titulo: 'Información adicional', Componente: Familiares, listo: (p) => p?.familiares?.length > 0 },
  { id: 'cargo', titulo: 'Cargo', Componente: Cargo, listo: (p) => !!p?.cargo },
  { id: 'formacion', titulo: 'Formación académica', Componente: Formacion, listo: (p) => p?.formacion?.length > 0 },
  { id: 'experiencia', titulo: 'Experiencia', Componente: Experiencia, listo: (p) => p?.experiencia?.length > 0 },
  { id: 'publicaciones', titulo: 'Publicaciones', Componente: Publicaciones, listo: (p) => p?.publicaciones?.length > 0 },
  { id: 'capacitacion', titulo: 'Capacitación', Componente: Capacitaciones, listo: (p) => p?.capacitaciones?.length > 0 },
  { id: 'inhabilidades', titulo: 'Inhabilidades', Componente: Inhabilidades, listo: (p) => !!p?.inhabilidades },
  { id: 'envio', titulo: 'Guardar y enviar', Componente: null, listo: (p) => p?.estado === 'ENVIADA' },
]

function Cierre({ mensaje }) {
  const { cerrarSesion } = useAuth()
  return (
    <section className="mx-auto max-w-xl rounded-lg bg-white p-6 shadow-sm" data-cierre>
      <h1 className="mb-3 text-xl font-semibold">Postulación cerrada</h1>
      <Alerta>{mensaje}</Alerta>
      <button type="button" onClick={() => cerrarSesion()} className="mt-4 rounded-md bg-emerald-700 px-4 py-2 font-medium text-white hover:bg-emerald-800">
        Salir
      </button>
    </section>
  )
}

// EF-12: formulario de postulación por secciones
export default function Postulacion() {
  const [carga, setCarga] = useState({ estado: 'cargando' }) // cargando | listo | error | cierre
  const [intento, setIntento] = useState(0)
  const [postulacion, setPostulacion] = useState(null) // último estado aceptado por el servidor
  const [activa, setActiva] = useState('personal')
  const [cierre, setCierre] = useState('')

  useEffect(() => {
    let vigente = true
    Promise.all([getPerfil(), getPeriodo(), getPostulacion()])
      .then(([perfil, periodo, guardada]) => {
        if (!vigente) return
        setPostulacion(guardada)
        setCarga({ estado: 'listo', perfil, periodo })
      })
      .catch((err) => {
        if (!vigente) return
        // V5 / 1.1.6: si la nacionalidad no es ecuatoriana, no se permite continuar y se cierra la postulación
        if (err.status === 403) setCarga({ estado: 'cierre', mensaje: err.message })
        else setCarga({ estado: 'error', mensaje: err.message })
      })
    return () => {
      vigente = false
    }
  }, [intento])

  // Guarda la postulación completa con los cambios de una sección (P4: el backend reemplaza todo en una transacción)
  const guardar = useCallback(
    async (cambios) => {
      const nueva = { ...postulacion, ...cambios }
      const { codigo } = await guardarPostulacion(nueva)
      setPostulacion({ ...nueva, codigo, estado: 'BORRADOR' })
    },
    [postulacion],
  )

  const onCierre = useCallback((mensaje) => setCierre(mensaje), [])
  const onEnviada = useCallback((codigo) => setPostulacion((p) => ({ ...p, codigo, estado: 'ENVIADA', acepta_terminos: true })), [])

  if (carga.estado === 'cargando') return <p className="text-slate-500">Cargando su postulación…</p>
  if (carga.estado === 'cierre') return <Cierre mensaje={carga.mensaje} />
  if (cierre) return <Cierre mensaje={cierre} />
  if (carga.estado === 'error') {
    return (
      <section className="mx-auto max-w-md rounded-lg bg-white p-6 shadow-sm">
        <Alerta>{carga.mensaje}</Alerta>
        <button type="button" onClick={() => setIntento((n) => n + 1)} className="mt-4 underline">
          Reintentar
        </button>
      </section>
    )
  }

  const { perfil, periodo } = carga
  const enviada = postulacion?.estado === 'ENVIADA'
  const soloLectura = !periodo.abierto || enviada
  const props = { postulacion, perfil, guardar, soloLectura, onCierre }

  return (
    <>
      <div className="flex flex-col gap-4 print:hidden">
        <header className="flex flex-wrap items-baseline justify-between gap-2">
          <h1 className="text-xl font-semibold">Formulario de postulación</h1>
          {postulacion?.codigo && (
            <p className="text-sm text-slate-600">
              Código: <strong data-codigo>{postulacion.codigo}</strong>
            </p>
          )}
        </header>

        {!periodo.abierto && (
          <Alerta>
            El período de postulación no está habilitado (del {fechaUI(periodo.inicio)} al {fechaUI(periodo.fin)}). Puede consultar su información, pero no modificarla.
          </Alerta>
        )}
        {enviada && <Alerta tipo="info">Su postulación fue enviada y ya no puede modificarse.</Alerta>}

        <nav aria-label="Secciones del formulario" className="flex flex-wrap gap-2">
          {SECCIONES.map((s) => (
            <button
              key={s.id}
              type="button"
              onClick={() => setActiva(s.id)}
              aria-current={activa === s.id ? 'step' : undefined}
              className="rounded-full border border-slate-300 px-3 py-1 text-sm aria-[current=step]:border-emerald-700 aria-[current=step]:bg-emerald-700 aria-[current=step]:text-white"
            >
              {s.listo(postulacion) && s.id !== 'personal' ? '✓ ' : ''}
              {s.titulo}
            </button>
          ))}
        </nav>

        {/* Todas montadas, solo una visible: cambiar de pestaña no pierde lo que no se guardó */}
        {SECCIONES.map(({ id, Componente }) => (
          <div key={id} hidden={activa !== id} data-seccion={id}>
            {Componente ? <Componente {...props} /> : <Envio postulacion={postulacion} guardar={guardar} soloLectura={soloLectura} enviada={enviada} onEnviada={onEnviada} />}
          </div>
        ))}

        <EliminarCuenta />
      </div>
      <Impresion perfil={perfil} postulacion={postulacion} />
    </>
  )
}
