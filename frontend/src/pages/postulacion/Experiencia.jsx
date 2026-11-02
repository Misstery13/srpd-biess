import { useState } from 'react'
import CampoPDF from '../../components/CampoPDF.jsx'
import { AreaTexto, Campo, Seleccion } from '../../components/Formulario.jsx'
import ListaEditable from '../../components/ListaEditable.jsx'
import { aFechaApi, AREAS, calcularDuracion, deFechaApi, LIMITE_GENERAL, sumarDuraciones, textoDuracion, TIPOS_EXPERIENCIA } from '../../utils/ef12.js'
import { fechaUI, hoy, nombreOpcion } from './comun.js'
import Seccion from './Seccion.jsx'

const nuevo = () => ({ tipo: '', cargo: '', institucion: '', fecha_inicio: '', fecha_fin: '', actual: false, area: '', funciones: '', documento: null })

function validar(f) {
  const e = {}
  if (!TIPOS_EXPERIENCIA.some(([c]) => c === f.tipo)) e.tipo = 'Seleccione el tipo de experiencia.'
  if (!f.cargo.trim()) e.cargo = 'Ingrese el cargo o puesto.'
  if (!f.institucion.trim()) e.institucion = 'Ingrese la institución o empresa.'
  if (!f.fecha_inicio) e.fecha_inicio = 'Ingrese la fecha de inicio.'
  else if (f.fecha_inicio > hoy()) e.fecha_inicio = 'La fecha de inicio no puede ser futura.'
  if (!f.actual) {
    if (!f.fecha_fin) e.fecha_fin = 'Ingrese la fecha de finalización o marque "Trabajo actual".'
    else if (f.fecha_fin > hoy()) e.fecha_fin = 'La fecha de finalización no puede ser futura.'
    else if (f.fecha_inicio && f.fecha_fin < f.fecha_inicio) e.fecha_fin = 'La fecha de finalización no puede ser anterior a la de inicio.' // V7
  }
  if (!AREAS.includes(f.area)) e.area = 'Seleccione el área de las actividades.'
  if (f.funciones.length > 1000) e.funciones = 'Máximo 1000 caracteres.' // V8
  if (!f.documento) e.documento = 'Cargue el documento de respaldo.'
  return e
}

const duracion = (f) => calcularDuracion(f.fecha_inicio, f.actual ? null : f.fecha_fin)

function FormExperiencia({ valor, onChange, errores }) {
  const set = (campo) => (v) => onChange({ ...valor, [campo]: v?.target ? v.target.value : v })
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Seleccion id="exp-tipo" etiqueta="Tipo de experiencia" opciones={TIPOS_EXPERIENCIA} required value={valor.tipo} onChange={set('tipo')} error={errores.tipo} />
      <Campo id="exp-cargo" etiqueta="Cargo o puesto" maxLength={150} required value={valor.cargo} onChange={set('cargo')} error={errores.cargo} />
      <Campo id="exp-institucion" etiqueta="Institución o empresa" maxLength={150} required value={valor.institucion} onChange={set('institucion')} error={errores.institucion} />
      <Seleccion id="exp-area" etiqueta="Actividades relacionadas a" opciones={AREAS} required value={valor.area} onChange={set('area')} error={errores.area} />
      <Campo id="exp-inicio" etiqueta="Fecha inicio" type="date" max={hoy()} required value={valor.fecha_inicio} onChange={set('fecha_inicio')} error={errores.fecha_inicio} />
      <div className="flex flex-col gap-2">
        <Campo id="exp-fin" etiqueta="Fecha finalización" type="date" max={hoy()} disabled={valor.actual} value={valor.actual ? '' : valor.fecha_fin} onChange={set('fecha_fin')} error={errores.fecha_fin} />
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={valor.actual} onChange={(e) => onChange({ ...valor, actual: e.target.checked, fecha_fin: '' })} />
          Trabajo actual
        </label>
      </div>
      <div className="sm:col-span-2">
        <AreaTexto id="exp-funciones" etiqueta="Descripción de funciones" maxLength={1000} value={valor.funciones} onChange={set('funciones')} error={errores.funciones} />
      </div>
      <p className="text-sm sm:col-span-2" data-duracion>
        Tiempo de experiencia: <strong>{textoDuracion(duracion(valor))}</strong>
      </p>
      <div className="sm:col-span-2">
        <CampoPDF id="exp-documento" etiqueta="Documento de respaldo" tipo="EXPERIENCIA" limite={LIMITE_GENERAL} obligatorio value={valor.documento} onChange={set('documento')} error={errores.documento} />
      </div>
    </div>
  )
}

const desdeApi = (f) => ({
  ...f,
  fecha_inicio: deFechaApi(f.fecha_inicio),
  fecha_fin: deFechaApi(f.fecha_fin),
  actual: !f.fecha_fin,
  funciones: f.funciones ?? '',
  documento: f.documento ?? null,
})
const aApi = (f) => ({
  tipo: f.tipo,
  cargo: f.cargo.trim(),
  institucion: f.institucion.trim(),
  fecha_inicio: aFechaApi(f.fecha_inicio),
  fecha_fin: f.actual ? null : aFechaApi(f.fecha_fin),
  area: f.area,
  funciones: f.funciones.trim() || null,
  documento: f.documento,
})

// 1.6: experiencia profesional, con la calculadora de tiempo por experiencia (años, meses y días)
export default function Experiencia({ postulacion, guardar, soloLectura }) {
  const [lista, setLista] = useState(() => (postulacion?.experiencia ?? []).map(desdeApi))
  const onGuardar = () => guardar({ experiencia: lista.map(aApi) })
  const total = lista.length ? sumarDuraciones(lista.map(duracion)) : null

  return (
    <Seccion titulo="Experiencia profesional" onGuardar={onGuardar} soloLectura={soloLectura} textoBoton="Guardar información">
      <ListaEditable
        nombre="experiencias"
        items={lista}
        onChange={setLista}
        nuevo={nuevo}
        validar={validar}
        Formulario={FormExperiencia}
        resumen={(f) => (
          <>
            <p className="font-medium">
              {f.cargo} · {f.institucion}
            </p>
            <p className="text-slate-500">
              {nombreOpcion(TIPOS_EXPERIENCIA, f.tipo)} · {f.area} · {fechaUI(f.fecha_inicio)} – {f.actual ? 'actualidad' : fechaUI(f.fecha_fin)}
            </p>
            <p data-duracion>
              Tiempo: <strong>{textoDuracion(duracion(f))}</strong>
            </p>
          </>
        )}
      />
      {total && (
        <p className="text-sm" data-total>
          Total de experiencia registrada: <strong>{textoDuracion(total)}</strong>
        </p>
      )}
    </Seccion>
  )
}
