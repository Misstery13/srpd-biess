import { useState } from 'react'
import CampoPDF from '../../components/CampoPDF.jsx'
import { Campo, Seleccion, SiNo } from '../../components/Formulario.jsx'
import ListaEditable from '../../components/ListaEditable.jsx'
import { aFechaApi, deFechaApi, LIMITE_GENERAL, TIPOS_PUBLICACION } from '../../utils/ef12.js'
import { fechaUI, hoy, nombreOpcion } from './comun.js'
import Seccion from './Seccion.jsx'

const nuevo = () => ({ titulo: '', tipo: '', medio: '', fecha: '', relacion_profesional: null, documento: null })

function validar(f) {
  const e = {}
  if (!f.titulo.trim()) e.titulo = 'Ingrese el título de la publicación o trabajo de investigación.'
  if (!TIPOS_PUBLICACION.some(([c]) => c === f.tipo)) e.tipo = 'Seleccione el tipo.'
  if (!f.medio.trim()) e.medio = 'Ingrese el medio de publicación.'
  if (!f.fecha) e.fecha = 'Ingrese la fecha de publicación.'
  else if (f.fecha > hoy()) e.fecha = 'La fecha no puede ser futura.'
  if (f.relacion_profesional === null) e.relacion_profesional = 'Responda Si o No.'
  if (!f.documento) e.documento = 'Cargue el documento de respaldo.'
  return e
}

function FormPublicacion({ valor, onChange, errores }) {
  const set = (campo) => (v) => onChange({ ...valor, [campo]: v?.target ? v.target.value : v })
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <div className="sm:col-span-2">
        <Campo id="pub-titulo" etiqueta="Título de la publicación o trabajo de investigación" maxLength={250} required value={valor.titulo} onChange={set('titulo')} error={errores.titulo} />
      </div>
      <Seleccion id="pub-tipo" etiqueta="Tipo" opciones={TIPOS_PUBLICACION} required value={valor.tipo} onChange={set('tipo')} error={errores.tipo} />
      <Campo id="pub-medio" etiqueta="Medio de publicación" maxLength={200} required value={valor.medio} onChange={set('medio')} error={errores.medio} />
      <Campo id="pub-fecha" etiqueta="Fecha de publicación" type="date" max={hoy()} required value={valor.fecha} onChange={set('fecha')} error={errores.fecha} />
      <SiNo id="pub-relacion" pregunta="¿Guarda relación con el campo de su ejercicio o formación profesional?" value={valor.relacion_profesional} onChange={set('relacion_profesional')} error={errores.relacion_profesional} />
      <div className="sm:col-span-2">
        <CampoPDF id="pub-documento" etiqueta="Documento de respaldo" tipo="PUBLICACION" limite={LIMITE_GENERAL} obligatorio value={valor.documento} onChange={set('documento')} error={errores.documento} />
      </div>
    </div>
  )
}

const desdeApi = (p) => ({ ...p, fecha: deFechaApi(p.fecha), documento: p.documento ?? null })
const aApi = (p) => ({ titulo: p.titulo.trim(), tipo: p.tipo, medio: p.medio.trim(), fecha: aFechaApi(p.fecha), relacion_profesional: p.relacion_profesional, documento: p.documento })

// 1.7: publicaciones e investigaciones académicas
export default function Publicaciones({ postulacion, guardar, soloLectura }) {
  const guardadas = postulacion?.publicaciones ?? []
  const [tiene, setTiene] = useState(guardadas.length ? true : null)
  const [lista, setLista] = useState(() => guardadas.map(desdeApi))
  const [error, setError] = useState('')

  async function onGuardar() {
    let msg = ''
    if (tiene === null) msg = 'Responda si cuenta con publicaciones o investigaciones.'
    else if (tiene && lista.length === 0) msg = 'Agregue al menos una publicación o responda "No".'
    setError(msg)
    if (msg) return { tiene: msg }
    await guardar({ publicaciones: tiene ? lista.map(aApi) : [] })
  }

  return (
    <Seccion titulo="Publicaciones" onGuardar={onGuardar} soloLectura={soloLectura} textoBoton="Guardar información">
      <SiNo id="tiene-publicaciones" pregunta="¿Cuenta con Publicaciones e Investigaciones académicas?" value={tiene} onChange={setTiene} error={error} />
      {tiene && (
        <ListaEditable
          nombre="publicaciones"
          items={lista}
          onChange={setLista}
          nuevo={nuevo}
          validar={validar}
          Formulario={FormPublicacion}
          resumen={(p) => (
            <>
              <p className="font-medium">{p.titulo}</p>
              <p className="text-slate-500">
                {nombreOpcion(TIPOS_PUBLICACION, p.tipo)} · {p.medio} · {fechaUI(p.fecha)} · Relación profesional: {p.relacion_profesional ? 'Si' : 'No'}
              </p>
            </>
          )}
        />
      )}
    </Seccion>
  )
}
