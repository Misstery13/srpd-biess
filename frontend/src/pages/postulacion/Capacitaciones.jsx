import { useState } from 'react'
import CampoPDF from '../../components/CampoPDF.jsx'
import { Campo, Seleccion } from '../../components/Formulario.jsx'
import ListaEditable from '../../components/ListaEditable.jsx'
import { aFechaApi, deFechaApi, LIMITE_GENERAL, TIPOS_EVENTO } from '../../utils/ef12.js'
import { fechaUI, hoy, nombreOpcion } from './comun.js'
import { fechaEnRango, revisarTexto, soloDigitos } from '../../utils/validaciones.js'
import Seccion from './Seccion.jsx'

const MAX_HORAS = 9999 // igual que en el backend (campos.go)

const nuevo = () => ({ tipo_evento: '', nombre_evento: '', institucion: '', fecha_inicio: '', fecha_fin: '', horas: '', documento: null })

function validar(c) {
  const e = {}
  if (!TIPOS_EVENTO.some(([v]) => v === c.tipo_evento)) e.tipo_evento = 'Seleccione el tipo de evento.'
  revisarTexto(e, 'nombre_evento', c.nombre_evento, { vacio: 'Ingrese el nombre del evento.' })
  revisarTexto(e, 'institucion', c.institucion, { vacio: 'Ingrese la institución.' })
  if (!c.fecha_inicio) e.fecha_inicio = 'Ingrese la fecha de inicio.'
  else if (!fechaEnRango(c.fecha_inicio, hoy())) e.fecha_inicio = 'Ingrese una fecha entre 1900 y hoy (no futura).'
  if (!c.fecha_fin) e.fecha_fin = 'Ingrese la fecha de fin.'
  else if (!fechaEnRango(c.fecha_fin, hoy())) e.fecha_fin = 'La fecha no puede ser futura.'
  else if (c.fecha_inicio && c.fecha_fin < c.fecha_inicio) e.fecha_fin = 'La fecha de fin no puede ser anterior a la de inicio.'
  if (!/^\d+$/.test(String(c.horas)) || Number(c.horas) <= 0 || Number(c.horas) > MAX_HORAS) e.horas = `Ingrese un número entero de horas entre 1 y ${MAX_HORAS}.`
  return e
}

function FormCapacitacion({ valor, onChange, errores }) {
  const set = (campo) => (v) => onChange({ ...valor, [campo]: v?.target ? v.target.value : v })
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Seleccion id="cap-tipo" etiqueta="Tipo de evento" opciones={TIPOS_EVENTO} required value={valor.tipo_evento} onChange={set('tipo_evento')} error={errores.tipo_evento} />
      <Campo id="cap-nombre" etiqueta="Nombre del evento" maxLength={200} required value={valor.nombre_evento} onChange={set('nombre_evento')} error={errores.nombre_evento} />
      <Campo id="cap-institucion" etiqueta="Institución" maxLength={150} required value={valor.institucion} onChange={set('institucion')} error={errores.institucion} />
      <Campo id="cap-horas" etiqueta="Número horas del evento" inputMode="numeric" maxLength={4} required value={valor.horas} onChange={(e) => set('horas')(soloDigitos(e.target.value))} error={errores.horas} />
      <Campo id="cap-inicio" etiqueta="Fecha inicio" type="date" min="1900-01-01" max={hoy()} required value={valor.fecha_inicio} onChange={set('fecha_inicio')} error={errores.fecha_inicio} />
      <Campo id="cap-fin" etiqueta="Fecha fin" type="date" min="1900-01-01" max={hoy()} required value={valor.fecha_fin} onChange={set('fecha_fin')} error={errores.fecha_fin} />
      <div className="sm:col-span-2">
        <CampoPDF id="cap-documento" etiqueta="Documento de respaldo" tipo="CAPACITACION" limite={LIMITE_GENERAL} value={valor.documento} onChange={set('documento')} />
      </div>
    </div>
  )
}

const desdeApi = (c) => ({ ...c, fecha_inicio: deFechaApi(c.fecha_inicio), fecha_fin: deFechaApi(c.fecha_fin), horas: String(c.horas), documento: c.documento ?? null })
const aApi = (c) => ({
  tipo_evento: c.tipo_evento,
  nombre_evento: c.nombre_evento.trim(),
  institucion: c.institucion.trim(),
  fecha_inicio: aFechaApi(c.fecha_inicio),
  fecha_fin: aFechaApi(c.fecha_fin),
  horas: Number(c.horas),
  documento: c.documento,
})

// 1.8: capacitación
// Lo que muestra cada tarjeta de la lista
const resumenCapacitaciones = (c) => (
  <>
    <p className="font-medium">{c.nombre_evento}</p>
    <p className="text-slate-500">
      {nombreOpcion(TIPOS_EVENTO, c.tipo_evento)} · {c.institucion} · {fechaUI(c.fecha_inicio)} – {fechaUI(c.fecha_fin)} · {c.horas} h
    </p>
  </>
)

export default function Capacitaciones({ postulacion, guardar, soloLectura }) {
  const [lista, setLista] = useState(() => (postulacion?.capacitaciones ?? []).map(desdeApi))
  // Sin registros ahora ni guardados antes no hay nada que guardar (vaciar una lista guardada sí es un cambio)
  const onGuardar = () =>
    lista.length === 0 && !postulacion?.capacitaciones?.length
      ? { _seccion: 'No ha registrado capacitaciones. Use "+ Agregar" para registrar.' }
      : guardar({ capacitaciones: lista.map(aApi) })

  return (
    <Seccion titulo="Capacitación" onGuardar={onGuardar} soloLectura={soloLectura} textoBoton="Guardar información">
      <ListaEditable
        nombre="capacitaciones"
        items={lista}
        onChange={setLista}
        nuevo={nuevo}
        validar={validar}
        Formulario={FormCapacitacion}
        resumen={resumenCapacitaciones}
      />
    </Seccion>
  )
}
