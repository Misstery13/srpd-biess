import { useEffect, useState } from 'react'
import { getCantones, getPaises, getProvincias, UBICACION_VACIA } from '../services/catalogos.js'

// Carga una lista cuando cambia `clave`; clave vacía = lista vacía sin pedir nada.
// "cargando" se deriva comparando la clave pedida con la última respondida: sin setState extra.
function useCatalogo(cargar, clave) {
  const [respuesta, setRespuesta] = useState({ clave: null, datos: [], error: null })
  const [intento, setIntento] = useState(0)

  useEffect(() => {
    if (!clave) return
    let vigente = true // ignora respuestas de una selección anterior
    cargar(clave)
      .then((datos) => vigente && setRespuesta({ clave, datos, error: null }))
      .catch((e) => vigente && setRespuesta({ clave, datos: [], error: e.message }))
    return () => {
      vigente = false
    }
  }, [cargar, clave, intento])

  const actual = !!clave && respuesta.clave === clave
  return {
    datos: actual ? respuesta.datos : [],
    error: actual ? respuesta.error : null,
    cargando: !!clave && !actual,
    reintentar: () => {
      setRespuesta((r) => ({ ...r, clave: null }))
      setIntento((n) => n + 1)
    },
  }
}

const campo = 'w-full rounded-md border border-slate-300 bg-white px-3 py-2 disabled:bg-slate-100 disabled:text-slate-400'

function Lista({ id, etiqueta, valor, onChange, catalogo, deshabilitado, placeholder }) {
  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium">
        {etiqueta} <span className="text-red-600">*</span>
      </label>
      <select
        id={id}
        className={campo}
        value={valor}
        onChange={(e) => onChange(e.target.value)}
        disabled={deshabilitado || catalogo.cargando || !!catalogo.error}
        required
      >
        <option value="">{catalogo.cargando ? 'Cargando…' : placeholder}</option>
        {catalogo.datos.map((item) => (
          <option key={item.id} value={item.id}>
            {item.nombre}
          </option>
        ))}
      </select>
      {catalogo.error && (
        <p className="mt-1 text-sm text-red-600" role="alert">
          {catalogo.error}{' '}
          <button type="button" onClick={catalogo.reintentar} className="underline">
            Reintentar
          </button>
        </p>
      )}
    </div>
  )
}

function Texto({ id, etiqueta, valor, onChange }) {
  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium">
        {etiqueta} <span className="text-red-600">*</span>
      </label>
      <input id={id} className={campo} value={valor} onChange={(e) => onChange(e.target.value)} maxLength={100} required />
    </div>
  )
}

// EF-12 1.2.1–1.2.3: País → Provincia → Ciudad; si el país no es Ecuador, provincia y ciudad son texto libre (V6)
export default function SelectorUbicacion({ value, onChange }) {
  const paises = useCatalogo(getPaises, 'todos')
  const pais = paises.datos.find((p) => String(p.id) === value.paisId)
  const esEcuador = pais?.nombre === 'Ecuador'
  // Solo Ecuador tiene catálogo de provincias y cantones
  const provincias = useCatalogo(getProvincias, esEcuador ? value.paisId : '')
  const cantones = useCatalogo(getCantones, value.provinciaId)

  return (
    <div className="grid gap-4 sm:grid-cols-3">
      <Lista
        id="pais"
        etiqueta="País"
        valor={value.paisId}
        onChange={(paisId) => onChange({ ...UBICACION_VACIA, paisId })}
        catalogo={paises}
        placeholder="Seleccione un país"
      />
      {!pais || esEcuador ? (
        <>
          <Lista
            id="provincia"
            etiqueta="Provincia"
            valor={value.provinciaId}
            onChange={(provinciaId) => onChange({ ...value, provinciaId, cantonId: '' })}
            catalogo={provincias}
            deshabilitado={!value.paisId}
            placeholder="Seleccione una provincia"
          />
          <Lista
            id="canton"
            etiqueta="Ciudad"
            valor={value.cantonId}
            onChange={(cantonId) => onChange({ ...value, cantonId })}
            catalogo={cantones}
            deshabilitado={!value.provinciaId}
            placeholder="Seleccione una ciudad"
          />
        </>
      ) : (
        <>
          <Texto id="provinciaTexto" etiqueta="Provincia / Estado" valor={value.provinciaTexto} onChange={(provinciaTexto) => onChange({ ...value, provinciaTexto })} />
          <Texto id="ciudadTexto" etiqueta="Ciudad" valor={value.ciudadTexto} onChange={(ciudadTexto) => onChange({ ...value, ciudadTexto })} />
        </>
      )}
    </div>
  )
}
