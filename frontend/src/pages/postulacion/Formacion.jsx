import { useState } from 'react'
import CampoPDF from '../../components/CampoPDF.jsx'
import { Campo, Seleccion } from '../../components/Formulario.jsx'
import ListaEditable from '../../components/ListaEditable.jsx'
import { aFechaApi, AREAS, deFechaApi, LIMITE_GENERAL, NIVELES, TIPOS_FORMACION } from '../../utils/ef12.js'
import { hoy, nombreOpcion } from './comun.js'
import { esSenescyt, fechaEnRango, revisarTexto } from '../../utils/validaciones.js'
import Seccion from './Seccion.jsx'

const nuevo = () => ({ nivel: '', institucion: '', tipo: '', titulo: '', area: '', senescyt_numero: '', senescyt_fecha: '', documento: null })

function validar(f) {
  const e = {}
  if (!NIVELES.includes(f.nivel)) e.nivel = 'Seleccione el nivel de instrucción.'
  revisarTexto(e, 'institucion', f.institucion, { vacio: 'Ingrese la institución educativa.' })
  if (!TIPOS_FORMACION.some(([c]) => c === f.tipo)) e.tipo = 'Seleccione el tipo.'
  revisarTexto(e, 'titulo', f.titulo, { vacio: 'Ingrese el título obtenido.' })
  if (!AREAS.includes(f.area)) e.area = 'Seleccione el área de estudios.'
  if (f.senescyt_numero.trim() && !esSenescyt(f.senescyt_numero.trim())) e.senescyt_numero = 'Use números, letras y guiones (p. ej. 1005-2018-1234567).'
  if (f.senescyt_fecha && !fechaEnRango(f.senescyt_fecha, hoy())) e.senescyt_fecha = 'Ingrese una fecha entre 1900 y hoy.'
  return e
}

function FormTitulo({ valor, onChange, errores }) {
  const set = (campo) => (v) => onChange({ ...valor, [campo]: v?.target ? v.target.value : v })
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Seleccion id="for-nivel" etiqueta="Nivel de instrucción" opciones={NIVELES} required value={valor.nivel} onChange={set('nivel')} error={errores.nivel} />
      <Campo id="for-institucion" etiqueta="Institución educativa" maxLength={150} required value={valor.institucion} onChange={set('institucion')} error={errores.institucion} />
      <Seleccion id="for-tipo" etiqueta="Tipo" opciones={TIPOS_FORMACION} required value={valor.tipo} onChange={set('tipo')} error={errores.tipo} />
      <Campo id="for-titulo" etiqueta="Título obtenido" maxLength={200} required value={valor.titulo} onChange={set('titulo')} error={errores.titulo} />
      <Seleccion id="for-area" etiqueta="Área de estudios" opciones={AREAS} required value={valor.area} onChange={set('area')} error={errores.area} />
      <Campo id="for-senescyt" etiqueta="Número de registro SENESCYT" maxLength={50} value={valor.senescyt_numero} onChange={set('senescyt_numero')} error={errores.senescyt_numero} />
      <Campo id="for-senescyt-fecha" etiqueta="Fecha de registro SENESCYT" type="date" min="1900-01-01" max={hoy()} value={valor.senescyt_fecha} onChange={set('senescyt_fecha')} error={errores.senescyt_fecha} />
      <CampoPDF id="for-documento" etiqueta="Documento de respaldo" tipo="FORMACION" limite={LIMITE_GENERAL} value={valor.documento} onChange={set('documento')} />
    </div>
  )
}

const desdeApi = (f) => ({ ...f, senescyt_numero: f.senescyt_numero ?? '', senescyt_fecha: deFechaApi(f.senescyt_fecha), documento: f.documento ?? null })
const aApi = (f) => ({
  nivel: f.nivel,
  institucion: f.institucion.trim(),
  tipo: f.tipo,
  titulo: f.titulo.trim(),
  area: f.area,
  senescyt_numero: f.senescyt_numero.trim() || null,
  senescyt_fecha: aFechaApi(f.senescyt_fecha),
  documento: f.documento,
})

// 1.5: formación académica
// Lo que muestra cada tarjeta de la lista
const resumenFormacion = (f) => (
  <>
    <p className="font-medium">{f.titulo}</p>
    <p className="text-slate-500">
      {f.nivel} · {f.institucion} ({nombreOpcion(TIPOS_FORMACION, f.tipo)}) · {f.area}
      {f.documento && ' · ✓ PDF'}
    </p>
  </>
)

export default function Formacion({ postulacion, guardar, soloLectura }) {
  const [lista, setLista] = useState(() => (postulacion?.formacion ?? []).map(desdeApi))
  // Sin registros ahora ni guardados antes no hay nada que guardar (vaciar una lista guardada sí es un cambio)
  const onGuardar = () =>
    lista.length === 0 && !postulacion?.formacion?.length
      ? { _seccion: 'No ha registrado títulos. Use "+ Agregar" para registrar.' }
      : guardar({ formacion: lista.map(aApi) })

  return (
    <Seccion titulo="Formación académica" onGuardar={onGuardar} soloLectura={soloLectura} textoBoton="Guardar información">
      <ListaEditable
        nombre="títulos"
        items={lista}
        onChange={setLista}
        nuevo={nuevo}
        validar={validar}
        Formulario={FormTitulo}
        resumen={resumenFormacion}
      />
    </Seccion>
  )
}
