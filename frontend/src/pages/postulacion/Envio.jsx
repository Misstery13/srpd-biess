import { useState } from 'react'
import { Alerta, BotonLinea } from '../../components/Formulario.jsx'
import { enviarPostulacion } from '../../services/postulacion.js'
import { CONSENTIMIENTOS } from '../../utils/ef12.js'
import { faltantes } from './comun.js'


const boton = 'rounded-md px-4 py-2 font-medium disabled:cursor-not-allowed disabled:bg-slate-400'

// 1.10: "Guardar Postulación" muestra los 5 textos de consentimiento (casillas sin marcar, L1); "Enviar Postulación"
export default function Envio({ postulacion, guardar, soloLectura, enviada, onEnviada }) {
  const [mostrarConsentimiento, setMostrarConsentimiento] = useState(false)
  const [aceptados, setAceptados] = useState(() => CONSENTIMIENTOS.map(() => false))
  const [estado, setEstado] = useState({ cargando: false, error: '' })
  const [aviso, setAviso] = useState('') // por qué no se envió (en una línea)
  const pendientes = faltantes(postulacion)

  async function ejecutar(accion) {
    setEstado({ cargando: true, error: '' })
    try {
      await accion()
      setEstado({ cargando: false, error: '' })
    } catch (err) {
      setEstado({ cargando: false, error: err.message })
    }
  }

  const guardarPostulacion = () =>
    ejecutar(async () => {
      await guardar({})
      setMostrarConsentimiento(true)
    })

  const enviar = () => {
    const faltan = aceptados.filter((a) => !a).length
    const motivo = pendientes.length
      ? `Falta guardar: ${pendientes.join(', ')}.`
      : faltan
        ? `Marque las ${CONSENTIMIENTOS.length} declaraciones (${faltan === 1 ? 'falta 1' : `faltan ${faltan}`}).`
        : ''
    setAviso(motivo)
    if (motivo) return
    return ejecutar(async () => {
      const { codigo } = await enviarPostulacion()
      onEnviada(codigo)
    })
  }

  return (
    <section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm sm:p-8">
      <h2 className="text-lg font-semibold text-marca-950">Guardar y enviar postulación</h2>

      {enviada ? (
        <div className="mt-4 flex flex-col gap-3">
          <Alerta tipo="info">
            Su postulación {postulacion.codigo} fue enviada. Se remitió el formulario a su correo electrónico. Ya no puede modificarse.
          </Alerta>
        </div>
      ) : (
        <fieldset disabled={soloLectura || estado.cargando} className="mt-4 flex min-w-0 flex-col gap-4">
          <Alerta>{estado.error}</Alerta>
          <div>
            <button type="button" onClick={guardarPostulacion} className={`${boton} bg-marca-700 text-white hover:bg-marca-800`}>
              Guardar Postulación
            </button>
          </div>

          {mostrarConsentimiento && (
            <div className="flex flex-col gap-3" data-consentimiento>
              <p className="text-sm text-slate-600">Lea y acepte cada declaración para poder enviar su postulación:</p>
              {CONSENTIMIENTOS.map((texto, i) => (
                <label key={texto} className="flex items-start gap-3 rounded-md border border-slate-200 p-3 text-sm">
                  <input
                    type="checkbox"
                    className="mt-1"
                    checked={aceptados[i]}
                    onChange={(e) => setAceptados(aceptados.map((a, j) => (j === i ? e.target.checked : a)))}
                  />
                  <span>
                    {i + 1}. {texto}
                  </span>
                </label>
              ))}
              <p className="text-sm text-slate-600">Una vez enviada, la postulación no podrá modificarse.</p>
              <div>
                <button
                  type="button"
                  onClick={enviar}
                  className={`${boton} bg-marca-700 text-white hover:bg-marca-800`}
                  data-enviar
                >
                  {estado.cargando ? 'Enviando…' : 'Enviar Postulación'}
                </button>
                {aviso && (
                  <p role="alert" className="mt-2 text-sm text-red-700" data-aviso>
                    {aviso}
                  </p>
                )}
              </div>
            </div>
          )}
        </fieldset>
      )}

      {/* 1.10.1 punto 5: se imprime y firma a mano lo que se envió, así que solo después de enviar */}
      <div className="mt-4 flex flex-wrap items-center gap-3 border-t border-slate-200 pt-4">
        <BotonLinea onClick={() => window.print()} disabled={!enviada}>
          Imprimir formulario
        </BotonLinea>
        {!enviada && <span className="text-sm text-slate-500">Disponible después de enviar la postulación.</span>}
      </div>
    </section>
  )
}
