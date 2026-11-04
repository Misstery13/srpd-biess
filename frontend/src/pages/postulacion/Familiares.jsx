import { useState } from 'react'
import { Campo, Seleccion, SiNo } from '../../components/Formulario.jsx'
import ListaEditable from '../../components/ListaEditable.jsx'
import { INSTITUCIONES, PARENTESCOS } from '../../utils/ef12.js'
import { revisarTexto, sinDigitos } from '../../utils/validaciones.js'
import Seccion from './Seccion.jsx'

const nuevo = () => ({ nombres: '', parentesco: '', institucion: '', area: '', cargo: '' })

function validar(f) {
  const e = {}
  revisarTexto(e, 'nombres', f.nombres, { vacio: 'Ingrese los nombres y apellidos del familiar.', formato: 'nombre' })
  if (!PARENTESCOS.includes(f.parentesco)) e.parentesco = 'Seleccione el parentesco.'
  if (!INSTITUCIONES.includes(f.institucion)) e.institucion = 'Seleccione la institución.'
  revisarTexto(e, 'area', f.area)
  revisarTexto(e, 'cargo', f.cargo)
  return e
}

function FormFamiliar({ valor, onChange, errores }) {
  const set = (campo) => (v) => onChange({ ...valor, [campo]: v?.target ? v.target.value : v })
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Campo id="fam-nombres" etiqueta="Nombres y apellidos del familiar" maxLength={200} required value={valor.nombres} onChange={(e) => set('nombres')(sinDigitos(e.target.value))} error={errores.nombres} />
      <Seleccion id="fam-parentesco" etiqueta="Parentesco familiar" opciones={PARENTESCOS} required value={valor.parentesco} onChange={set('parentesco')} error={errores.parentesco} />
      <Seleccion id="fam-institucion" etiqueta="Institución donde trabaja el familiar" opciones={INSTITUCIONES} required value={valor.institucion} onChange={set('institucion')} error={errores.institucion} />
      <Campo id="fam-area" etiqueta="Área donde trabaja el familiar" maxLength={100} value={valor.area ?? ''} onChange={set('area')} error={errores.area} />
      <Campo id="fam-cargo" etiqueta="Cargo" maxLength={100} value={valor.cargo ?? ''} onChange={set('cargo')} error={errores.cargo} />
    </div>
  )
}

const limpio = (f) => ({ nombres: f.nombres.trim(), parentesco: f.parentesco, institucion: f.institucion, area: f.area?.trim() || null, cargo: f.cargo?.trim() || null })

// 1.3: información adicional — familiares en el IESS o BIESS
// Lo que muestra cada tarjeta de la lista
const resumenFamiliares = (f) => (
  <>
    <p className="font-medium">{f.nombres}</p>
    <p className="text-slate-500">
      {f.parentesco} · {f.institucion}
      {f.area && ` · ${f.area}`}
      {f.cargo && ` · ${f.cargo}`}
    </p>
  </>
)

export default function Familiares({ postulacion, guardar, soloLectura }) {
  const guardados = postulacion?.familiares ?? []
  const [tiene, setTiene] = useState(guardados.length ? true : null)
  const [lista, setLista] = useState(guardados)
  const [error, setError] = useState('')

  async function onGuardar() {
    let msg = ''
    if (tiene === null) msg = 'Responda si tiene familiares trabajando en el IESS o BIESS.'
    else if (tiene && lista.length === 0) msg = 'Agregue al menos un familiar o responda "No".'
    setError(msg)
    if (msg) return { tiene: msg }
    await guardar({ familiares: tiene ? lista.map(limpio) : [] })
  }

  return (
    <Seccion titulo="Información adicional" onGuardar={onGuardar} soloLectura={soloLectura} textoBoton="Guardar información">
      <SiNo
        id="tiene-familiares"
        pregunta="Tiene familiares dentro del segundo grado de consanguinidad y cuarto de afinidad, trabajando en el IESS o BIESS actualmente"
        value={tiene}
        onChange={setTiene}
        error={error}
      />
      {tiene && (
        <ListaEditable
          nombre="familiares"
          items={lista}
          onChange={setLista}
          nuevo={nuevo}
          validar={validar}
          Formulario={FormFamiliar}
          resumen={resumenFamiliares}
        />
      )}
    </Seccion>
  )
}
