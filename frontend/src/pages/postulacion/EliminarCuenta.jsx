import { useState } from 'react'
import { Alerta, BotonLinea, Campo } from '../../components/Formulario.jsx'
import { useAuth } from '../../context/auth.js'
import { eliminarCuenta } from '../../services/postulacion.js'

const CONFIRMACION = 'ELIMINAR'

// L5: derecho de supresión. Requiere confirmación explícita (escribir ELIMINAR).
export default function EliminarCuenta() {
  const { cerrarSesion } = useAuth()
  const [abierto, setAbierto] = useState(false)
  const [texto, setTexto] = useState('')
  const [estado, setEstado] = useState({ cargando: false, error: '' })

  async function eliminar() {
    setEstado({ cargando: true, error: '' })
    try {
      await eliminarCuenta()
      // RutaProtegida lleva a /login, que muestra este aviso
      cerrarSesion('Su cuenta y sus datos personales fueron eliminados.')
    } catch (err) {
      setEstado({ cargando: false, error: err.message })
    }
  }

  return (
    <section className="rounded-lg border border-red-200 bg-white p-6 shadow-sm print:hidden">
      <h2 className="text-lg font-semibold text-red-800">Eliminar mi cuenta</h2>
      <p className="mt-1 text-sm text-slate-600">
        Se eliminarán su cuenta, su postulación, los datos de terceros que registró y los documentos cargados. Esta acción no se puede deshacer.
      </p>
      {abierto ? (
        <div className="mt-4 flex flex-col gap-3">
          <Alerta>{estado.error}</Alerta>
          <Campo id="confirmar-eliminar" etiqueta={`Para confirmar, escriba ${CONFIRMACION}`} autoComplete="off" value={texto} onChange={(e) => setTexto(e.target.value)} />
          <div className="flex gap-2">
            <BotonLinea peligro onClick={eliminar} disabled={texto !== CONFIRMACION || estado.cargando}>
              {estado.cargando ? 'Eliminando…' : 'Eliminar definitivamente'}
            </BotonLinea>
            <BotonLinea onClick={() => setAbierto(false)}>Cancelar</BotonLinea>
          </div>
        </div>
      ) : (
        <div className="mt-4">
          <BotonLinea peligro onClick={() => setAbierto(true)}>
            Eliminar mi cuenta
          </BotonLinea>
        </div>
      )}
    </section>
  )
}
