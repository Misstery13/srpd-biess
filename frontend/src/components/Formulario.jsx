// Piezas comunes de los formularios de EF-11 (E5: errores junto al campo o arriba, con role="alert")
const estiloCampo =
  'w-full rounded-md border border-slate-300 bg-white px-3 py-2 aria-invalid:border-red-600 disabled:bg-slate-100'

export function Campo({ id, etiqueta, error, ...props }) {
  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium">
        {etiqueta}
      </label>
      <input
        id={id}
        name={id}
        className={estiloCampo}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? `${id}-error` : undefined}
        {...props}
      />
      {error && (
        <p id={`${id}-error`} role="alert" className="mt-1 text-sm text-red-700">
          {error}
        </p>
      )}
    </div>
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

export function Tarjeta({ titulo, children }) {
  return (
    <section className="mx-auto max-w-md rounded-lg bg-white p-6 shadow-sm">
      <h1 className="mb-4 text-xl font-semibold">{titulo}</h1>
      {children}
    </section>
  )
}
