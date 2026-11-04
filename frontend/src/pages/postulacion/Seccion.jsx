import { useState } from 'react'
import { Alerta, Boton } from '../../components/Formulario.jsx'

// Contenedor de cada sección: título, "Guardar progreso" y mensajes (E5).
// onGuardar() devuelve { campo: mensaje } si hay errores de validación, o persiste y no devuelve nada.
// soloLectura deshabilita todo con <fieldset disabled> (período cerrado o postulación enviada).
export default function Seccion({ titulo, descripcion, onGuardar, soloLectura, textoBoton = 'Guardar progreso', children }) {
  const [estado, setEstado] = useState({ cargando: false, error: '', ok: false })

  async function enviar(e) {
    e.preventDefault()
    // Un ítem abierto en ListaEditable (sin "Agregar a la lista") no se guardaría: se perdería o quedaría vacío
    if (e.currentTarget.querySelector('[data-edicion]')) {
      setEstado({ cargando: false, error: 'Tiene un registro abierto: pulse "Agregar a la lista" (o "Aplicar cambios") o "Cancelar" antes de guardar.', ok: false })
      return
    }
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
    <section className="rounded-xl border border-slate-200 bg-white shadow-sm">
      <div className="border-b border-slate-200 px-5 py-4 sm:px-8">
        <h2 className="text-lg font-semibold text-marca-950">{titulo}</h2>
        {descripcion && <p className="mt-0.5 text-sm text-slate-500">{descripcion}</p>}
      </div>
      <form onSubmit={enviar} noValidate className="px-5 py-6 sm:px-8">
        <fieldset disabled={soloLectura} className="flex min-w-0 flex-col gap-5">
          <Alerta>{estado.error}</Alerta>
          {children}
          {onGuardar && (
            <div className="mt-2 flex flex-wrap items-center gap-3 border-t border-slate-100 pt-5">
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
