// Piezas comunes de los formularios (E5: errores junto al campo o arriba, con role="alert")
const estiloCampo =
  'w-full rounded-md border border-slate-300 bg-white px-3 py-2 aria-invalid:border-red-600 disabled:bg-slate-100'

// Etiqueta + control + error; (*) marca los obligatorios como en el documento
function Envoltura({ id, etiqueta, error, obligatorio, ayuda, children }) {
  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium">
        {etiqueta}
        {obligatorio && <span className="text-red-600"> *</span>}
      </label>
      {children}
      {ayuda && <p className="mt-1 text-xs text-slate-500">{ayuda}</p>}
      {error && (
        <p id={`${id}-error`} role="alert" className="mt-1 text-sm text-red-700">
          {error}
        </p>
      )}
    </div>
  )
}

const aria = (id, error) => ({
  id,
  name: id,
  className: estiloCampo,
  'aria-invalid': error ? true : undefined,
  'aria-describedby': error ? `${id}-error` : undefined,
})

export function Campo({ id, etiqueta, error, ayuda, ...props }) {
  return (
    <Envoltura id={id} etiqueta={etiqueta} error={error} ayuda={ayuda} obligatorio={props.required}>
      <input {...aria(id, error)} {...props} />
    </Envoltura>
  )
}

// opciones: ['Texto', ...] o [['CODIGO', 'Texto'], ...]
export function Seleccion({ id, etiqueta, error, opciones, value, onChange, required }) {
  return (
    <Envoltura id={id} etiqueta={etiqueta} error={error} obligatorio={required}>
      <select {...aria(id, error)} value={value ?? ''} onChange={(e) => onChange(e.target.value)} required={required}>
        <option value="">Seleccione…</option>
        {opciones.map((o) => {
          const [valor, texto] = Array.isArray(o) ? o : [o, o]
          return (
            <option key={valor} value={valor}>
              {texto}
            </option>
          )
        })}
      </select>
    </Envoltura>
  )
}

export function AreaTexto({ id, etiqueta, error, value, onChange, maxLength }) {
  return (
    <Envoltura id={id} etiqueta={etiqueta} error={error} ayuda={`${(value ?? '').length} / ${maxLength} caracteres`}>
      <textarea {...aria(id, error)} rows={4} maxLength={maxLength} value={value ?? ''} onChange={(e) => onChange(e.target.value)} />
    </Envoltura>
  )
}

// Opciones "Si/No" del documento. value: true | false | null (sin responder)
export function SiNo({ id, pregunta, value, onChange, error, obligatorio = true }) {
  return (
    <fieldset aria-describedby={error ? `${id}-error` : undefined}>
      <legend className="mb-1 text-sm font-medium">
        {pregunta}
        {obligatorio && <span className="text-red-600"> *</span>}
      </legend>
      <div className="flex gap-6 text-sm">
        {[
          [true, 'Si'],
          [false, 'No'],
        ].map(([valor, texto]) => (
          <label key={texto} className="flex items-center gap-2">
            <input type="radio" name={id} checked={value === valor} onChange={() => onChange(valor)} />
            {texto}
          </label>
        ))}
      </div>
      {error && (
        <p id={`${id}-error`} role="alert" className="mt-1 text-sm text-red-700">
          {error}
        </p>
      )}
    </fieldset>
  )
}

// first-letter:uppercase: el texto del DOM queda exactamente como el documento, solo cambia cómo se ve
export function Alerta({ tipo = 'error', children }) {
  if (!children) return null
  const estilo = 'block rounded-md border px-3 py-2 text-sm first-letter:uppercase'
  if (tipo === 'error') {
    return (
      <p role="alert" className={`${estilo} border-red-300 bg-red-50 text-red-800`}>
        {children}
      </p>
    )
  }
  // <output> tiene role="status" implícito: los lectores de pantalla anuncian el aviso
  return <output className={`${estilo} border-emerald-300 bg-emerald-50 text-emerald-900`}>{children}</output>
}

export function Boton({ cargando, disabled, children }) {
  return (
    <button
      type="submit"
      className="w-full rounded-md bg-emerald-700 px-4 py-2 font-medium text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-400"
      disabled={cargando || disabled}
    >
      {cargando ? 'Procesando…' : children}
    </button>
  )
}

// Botones secundarios de las listas (Agregar, Editar, Eliminar, Cancelar)
export function BotonLinea({ children, onClick, peligro = false, disabled }) {
  const color = peligro ? 'border-red-300 text-red-700 hover:bg-red-50' : 'border-emerald-700 text-emerald-800 hover:bg-emerald-50'
  return (
    <button type="button" onClick={onClick} disabled={disabled} className={`rounded-md border px-3 py-1.5 text-sm font-medium disabled:opacity-50 ${color}`}>
      {children}
    </button>
  )
}

export function Tarjeta({ titulo, children }) {
  return (
    <section className="mx-auto max-w-md rounded-lg bg-white p-6 shadow-sm">
      <h1 className="mb-4 text-xl font-semibold">{titulo}</h1>
      {children}
    </section>
  )
}
