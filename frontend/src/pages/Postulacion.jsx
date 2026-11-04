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
  { id: 'adicional', titulo: 'Información adicional', Componente: Familiares, listo: (p) => p?.tiene_familiares != null || p?.familiares?.length > 0 },
  { id: 'cargo', titulo: 'Cargo', Componente: Cargo, listo: (p) => !!p?.cargo },
  { id: 'formacion', titulo: 'Formación académica', Componente: Formacion, listo: (p) => p?.formacion?.length > 0 },
  { id: 'experiencia', titulo: 'Experiencia', Componente: Experiencia, listo: (p) => p?.experiencia?.length > 0 },
  { id: 'publicaciones', titulo: 'Publicaciones', Componente: Publicaciones, listo: (p) => p?.tiene_publicaciones != null || p?.publicaciones?.length > 0 },
  { id: 'capacitacion', titulo: 'Capacitación', Componente: Capacitaciones, listo: (p) => p?.capacitaciones?.length > 0 },
  { id: 'inhabilidades', titulo: 'Inhabilidades', Componente: Inhabilidades, listo: (p) => !!p?.inhabilidades },
  { id: 'envio', titulo: 'Guardar y enviar', Componente: null, listo: (p) => p?.estado === 'ENVIADA' },
]

function Cierre({ mensaje }) {
  const { cerrarSesion } = useAuth()
  return (
    <section className="mx-auto max-w-xl rounded-xl border border-slate-200 bg-white p-6 shadow-sm sm:p-8" data-cierre>
      <h1 className="mb-4 text-2xl font-semibold text-marca-950">Postulación cerrada</h1>
      <Alerta>{mensaje}</Alerta>
      <button type="button" onClick={() => cerrarSesion()} className="mt-6 rounded-md bg-marca-600 px-5 py-2.5 font-semibold text-white shadow-sm hover:bg-marca-700">
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
  const props = { postulacion, perfil, guardar, soloLectura, onCierre, irA: setActiva }

  return (
    <>
      <div className="flex flex-col gap-6 print:hidden">
        <header className="flex flex-wrap items-end justify-between gap-3 border-b border-slate-200 pb-5">
          <div>
            <p className="text-xs font-semibold uppercase tracking-widest text-marca-600">Postulación en línea</p>
            <h1 className="text-2xl font-semibold text-marca-950">Formulario de postulación</h1>
            <p className="mt-1 text-sm text-slate-500">
              {perfil.nombres} {perfil.apellidos} · C.I. {perfil.cedula}
            </p>
          </div>
          {postulacion?.codigo && (
            <p className="rounded-md border border-marca-200 bg-white px-3 py-2 text-sm text-slate-600">
              Código de postulación <strong data-codigo className="ml-1 font-mono text-marca-800">{postulacion.codigo}</strong>
            </p>
          )}
        </header>

        {!periodo.abierto && (
          <Alerta>
            El período de postulación no está habilitado (del {fechaUI(periodo.inicio)} al {fechaUI(periodo.fin)}). Puede consultar su información, pero no modificarla.
          </Alerta>
        )}
        {enviada && <Alerta tipo="info">Su postulación fue enviada y ya no puede modificarse.</Alerta>}

        <div className="grid gap-6 lg:grid-cols-[15rem_1fr] lg:items-start">
          {/* Celular: pestañas con desplazamiento horizontal. Computadora: menú lateral de pasos (número por CSS, fuera del texto) */}
          <nav
            aria-label="Secciones del formulario"
            className="-mx-4 flex gap-2 overflow-x-auto px-4 pb-1 [counter-reset:paso] lg:sticky lg:top-6 lg:mx-0 lg:flex-col lg:gap-1 lg:overflow-visible lg:rounded-xl lg:border lg:border-slate-200 lg:bg-white lg:p-3"
          >
            {SECCIONES.map((s) => (
              <button
                key={s.id}
                type="button"
                onClick={() => setActiva(s.id)}
                aria-current={activa === s.id ? 'step' : undefined}
                className="flex shrink-0 items-center gap-2 whitespace-nowrap rounded-full border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-700 [counter-increment:paso] before:hidden before:size-6 before:shrink-0 before:items-center before:justify-center before:rounded-full before:bg-slate-100 before:text-xs before:font-semibold before:text-slate-500 before:content-[counter(paso)] hover:border-marca-400 aria-[current=step]:border-marca-600 aria-[current=step]:bg-marca-600 aria-[current=step]:text-white lg:rounded-md lg:border-transparent lg:py-2 lg:text-left lg:before:flex lg:aria-[current=step]:bg-marca-50 lg:aria-[current=step]:font-medium lg:aria-[current=step]:text-marca-800 lg:aria-[current=step]:before:bg-marca-600 lg:aria-[current=step]:before:text-white"
              >
                {s.listo(postulacion) && s.id !== 'personal' ? '✓ ' : ''}
                {s.titulo}
              </button>
            ))}
          </nav>

          <div className="flex min-w-0 flex-col gap-6">
            {/* Todas montadas, solo una visible: cambiar de pestaña no pierde lo que no se guardó */}
            {SECCIONES.map(({ id, Componente }) => (
              <div key={id} hidden={activa !== id} data-seccion={id}>
                {Componente ? <Componente {...props} /> : <Envio postulacion={postulacion} guardar={guardar} soloLectura={soloLectura} enviada={enviada} onEnviada={onEnviada} />}
              </div>
            ))}
            <EliminarCuenta />
          </div>
        </div>
      </div>
      {enviada ? (
        <Impresion perfil={perfil} postulacion={postulacion} />
      ) : (
        <p className="hidden print:block">El formulario se imprime después de enviar la postulación.</p>
      )}
    </>
  )
}
