import { useState } from 'react'
import { Campo, SiNo } from '../../components/Formulario.jsx'
import { aFechaApi, deFechaApi, INHABILIDADES } from '../../utils/ef12.js'
import { hoy } from './comun.js'
import Seccion from './Seccion.jsx'

// Declaración jurada: ninguna respuesta viene marcada; las 13 son obligatorias
const vacio = () => ({
  ...Object.fromEntries(INHABILIDADES.map(([clave]) => [clave, null])),
  removido_institucion: '',
  removido_cargo: '',
  removido_fecha_fin: '',
  removido_motivo: '',
})

const desdeApi = (i) =>
  i
    ? { ...i, removido_institucion: i.removido_institucion ?? '', removido_cargo: i.removido_cargo ?? '', removido_fecha_fin: deFechaApi(i.removido_fecha_fin), removido_motivo: i.removido_motivo ?? '' }
    : vacio()

function aApi(f) {
  const respuestas = Object.fromEntries(INHABILIDADES.map(([clave]) => [clave, f[clave]]))
  const r = f.removido_control
  return {
    ...respuestas,
    removido_institucion: r ? f.removido_institucion.trim() : null,
    removido_cargo: r ? f.removido_cargo.trim() || null : null,
    removido_fecha_fin: r ? aFechaApi(f.removido_fecha_fin) : null,
    removido_motivo: r ? f.removido_motivo.trim() : null,
  }
}

function validar(f) {
  const e = {}
  for (const [clave] of INHABILIDADES) if (f[clave] === null) e[clave] = 'Seleccione Si o No.'
  if (f.removido_control) {
    if (!f.removido_institucion.trim()) e.removido_institucion = 'Ingrese la institución.'
    if (!f.removido_motivo.trim()) e.removido_motivo = 'Ingrese el motivo de finalización.'
    if (f.removido_fecha_fin > hoy()) e.removido_fecha_fin = 'La fecha no puede ser futura.'
  }
  return e
}

// 1.9: inhabilidades (13 preguntas)
export default function Inhabilidades({ postulacion, guardar, soloLectura }) {
  const [form, setForm] = useState(() => desdeApi(postulacion?.inhabilidades))
  const [errores, setErrores] = useState({})
  const set = (clave) => (v) => setForm({ ...form, [clave]: v?.target ? v.target.value : v })

  async function onGuardar() {
    const e = validar(form)
    setErrores(e)
    if (Object.keys(e).length) return e
    await guardar({ inhabilidades: aApi(form) })
  }

  return (
    <Seccion titulo="Inhabilidades" descripcion="Todas las preguntas son obligatorias." onGuardar={onGuardar} soloLectura={soloLectura} textoBoton="Guardar información">
      <ol className="flex flex-col gap-5">
        {INHABILIDADES.map(([clave, pregunta], i) => (
          <li key={clave} className="flex flex-col gap-3">
            <SiNo id={clave} pregunta={`${i + 1}. ${pregunta}`} value={form[clave]} onChange={set(clave)} error={errores[clave]} />
            {clave === 'removido_control' && form.removido_control && (
              <div className="grid gap-4 rounded-md border border-slate-200 p-4 sm:grid-cols-2">
                <Campo id="removido_institucion" etiqueta="Institución:" maxLength={200} required value={form.removido_institucion} onChange={set('removido_institucion')} error={errores.removido_institucion} />
                <Campo id="removido_cargo" etiqueta="Cargo:" maxLength={100} value={form.removido_cargo} onChange={set('removido_cargo')} />
                <Campo id="removido_fecha_fin" etiqueta="Fecha finalización:" type="date" max={hoy()} value={form.removido_fecha_fin} onChange={set('removido_fecha_fin')} error={errores.removido_fecha_fin} />
                <Campo id="removido_motivo" etiqueta="Motivo finalización:" maxLength={500} required value={form.removido_motivo} onChange={set('removido_motivo')} error={errores.removido_motivo} />
              </div>
            )}
          </li>
        ))}
      </ol>
    </Seccion>
  )
}
