import { useState } from 'react'
import { BotonLinea } from './Formulario.jsx'

// Clave estable por ítem para React (los ítems de la API no traen id propio)
const ids = new WeakMap()
let siguiente = 0
function idDe(item) {
  if (!ids.has(item)) ids.set(item, ++siguiente)
  return ids.get(item)
}

// Lista con "Agregar", "Editar" y "Eliminar" (EF-12: familiares, formación, experiencia, publicaciones, capacitación).
// Los cambios quedan en la lista local; la sección los persiste con "Guardar información".
//   items / onChange: la lista
//   nuevo(): ítem vacío
//   validar(item) → { campo: mensaje }
//   Formulario: ({ valor, onChange, errores }) => campos del ítem
//   resumen(item) → lo que se muestra en la tarjeta
export default function ListaEditable({ nombre, items, onChange, nuevo, validar, Formulario, resumen }) {
  const [edicion, setEdicion] = useState(null) // null | { indice: number | -1, valor, errores }

  const abrir = (indice) => setEdicion({ indice, valor: indice === -1 ? nuevo() : structuredClone(items[indice]), errores: {} })

  function aceptar() {
    const errores = validar(edicion.valor)
    if (Object.keys(errores).length) return setEdicion({ ...edicion, errores })
    onChange(edicion.indice === -1 ? [...items, edicion.valor] : items.map((it, i) => (i === edicion.indice ? edicion.valor : it)))
    setEdicion(null)
  }

  return (
    <div className="flex flex-col gap-3">
      {items.length === 0 && !edicion && <p className="text-sm text-slate-500">No ha registrado {nombre}.</p>}
      <ul className="flex flex-col gap-2">
        {items.map((item, i) => (
          <li key={idDe(item)} className="flex flex-wrap items-start justify-between gap-2 rounded-md border border-slate-200 p-3">
            <div className="min-w-0 flex-1 text-sm">{resumen(item)}</div>
            <div className="flex gap-2">
              <BotonLinea onClick={() => abrir(i)} disabled={!!edicion}>
                Editar
              </BotonLinea>
              <BotonLinea peligro onClick={() => onChange(items.filter((_, j) => j !== i))} disabled={!!edicion}>
                Eliminar
              </BotonLinea>
            </div>
          </li>
        ))}
      </ul>

      {edicion ? (
        <div className="flex flex-col gap-4 rounded-md border border-marca-200 bg-marca-50/40 p-4" data-edicion>
          <Formulario valor={edicion.valor} onChange={(valor) => setEdicion({ ...edicion, valor })} errores={edicion.errores} />
          <div className="flex gap-2">
            <BotonLinea onClick={aceptar}>{edicion.indice === -1 ? 'Agregar a la lista' : 'Aplicar cambios'}</BotonLinea>
            <BotonLinea peligro onClick={() => setEdicion(null)}>
              Cancelar
            </BotonLinea>
          </div>
        </div>
      ) : (
        <div>
          <BotonLinea onClick={() => abrir(-1)}>+ Agregar</BotonLinea>
        </div>
      )}
    </div>
  )
}
