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

  const enviar = () =>
    ejecutar(async () => {
      const { codigo } = await enviarPostulacion()
      onEnviada(codigo)
    })

  return (
    <section className="rounded-lg bg-white p-6 shadow-sm">
      <h2 className="text-lg font-semibold">Guardar y enviar postulación</h2>

      {enviada ? (
        <div className="mt-4 flex flex-col gap-3">
          <Alerta tipo="info">
            Su postulación {postulacion.codigo} fue enviada. Se remitió el formulario a su correo electrónico. Ya no puede modificarse.
          </Alerta>
        </div>
      ) : (
        <fieldset disabled={soloLectura || estado.cargando} className="mt-4 flex min-w-0 flex-col gap-4">
          <Alerta>{estado.error}</Alerta>
          {pendientes.length > 0 && (
            <div className="rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900" data-pendientes>
              <p className="font-medium">Para enviar la postulación falta guardar:</p>
              <ul className="mt-1 list-disc pl-5">
                {pendientes.map((p) => (
                  <li key={p}>{p}</li>
                ))}
              </ul>
            </div>
          )}

          <div>
            <button type="button" onClick={guardarPostulacion} className={`${boton} bg-emerald-700 text-white hover:bg-emerald-800`}>
              Guardar Postulación
            </button>
          </div>

          {mostrarConsentimiento && (
            <div className="flex flex-col gap-3" data-consentimiento>
              <p className="text-sm text-slate-600">Lea y acepte cada declaración para poder enviar su postulación:</p>
              {CONSENTIMIENTOS.map((texto, i) => (
                <label key={i} className="flex items-start gap-3 rounded-md border border-slate-200 p-3 text-sm">
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
                  disabled={!aceptados.every(Boolean) || pendientes.length > 0}
                  className={`${boton} bg-emerald-700 text-white hover:bg-emerald-800`}
                  data-enviar
                >
                  {estado.cargando ? 'Enviando…' : 'Enviar Postulación'}
                </button>
              </div>
            </div>
          )}
        </fieldset>
      )}

      {/* 1.10.1 punto 5: el formulario se imprime y se firma a mano */}
      <div className="mt-4 border-t border-slate-200 pt-4">
        <BotonLinea onClick={() => window.print()}>Imprimir formulario</BotonLinea>
      </div>
    </section>
  )
}
