import { useState } from 'react'
import { Alerta, Boton } from '../../components/Formulario.jsx'

// Contenedor de cada sección: título, "Guardar progreso" y mensajes (E5).
// onGuardar() devuelve { campo: mensaje } si hay errores de validación, o persiste y no devuelve nada.
// soloLectura deshabilita todo con <fieldset disabled> (período cerrado o postulación enviada).
export default function Seccion({ titulo, descripcion, onGuardar, soloLectura, textoBoton = 'Guardar progreso', children }) {
  const [estado, setEstado] = useState({ cargando: false, error: '', ok: false })

  async function enviar(e) {
    e.preventDefault()
    setEstado({ cargando: true, error: '', ok: false })
    try {
      const errores = await onGuardar()
      if (errores && Object.keys(errores).length) {
        setEstado({ cargando: false, error: 'Revise los campos marcados.', ok: false })
        return
      }
      setEstado({ cargando: false, error: '', ok: true })
    } catch (err) {
      setEstado({ cargando: false, error: err.message, ok: false })
    }
  }

  return (
    <section className="rounded-lg bg-white p-6 shadow-sm">
      <h2 className="text-lg font-semibold">{titulo}</h2>
      {descripcion && <p className="mt-1 text-sm text-slate-500">{descripcion}</p>}
      <form onSubmit={enviar} noValidate className="mt-4">
        <fieldset disabled={soloLectura} className="flex min-w-0 flex-col gap-4">
          <Alerta>{estado.error}</Alerta>
          {children}
          {onGuardar && (
            <div className="flex flex-wrap items-center gap-3 pt-2">
              <div className="w-full sm:w-56">
                <Boton cargando={estado.cargando}>{textoBoton}</Boton>
              </div>
              {estado.ok && <Alerta tipo="info">Información guardada.</Alerta>}
            </div>
          )}
        </fieldset>
      </form>
    </section>
  )
}
