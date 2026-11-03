// Piezas comunes de los formularios (E5: errores junto al campo o arriba, con role="alert")
const estiloCampo =
  'w-full rounded-md border border-slate-300 bg-white px-3 py-2.5 text-slate-900 shadow-xs transition-colors placeholder:text-slate-400 hover:border-slate-400 focus:border-marca-600 focus:outline-none focus:ring-3 focus:ring-marca-600/15 aria-invalid:border-red-600 aria-invalid:focus:ring-red-600/15 disabled:bg-slate-100 disabled:text-slate-500'

// Etiqueta + control + error; (*) marca los obligatorios como en el documento
function Envoltura({ id, etiqueta, error, obligatorio, ayuda, children }) {
  return (
    <div>
      <label htmlFor={id} className="mb-1.5 block text-sm font-medium text-slate-700">
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
  const estilo = 'block rounded-md border border-l-4 px-3 py-2.5 text-sm first-letter:uppercase'
  if (tipo === 'error') {
    return (
      <p role="alert" className={`${estilo} border-red-200 border-l-red-600 bg-red-50 text-red-800`}>
        {children}
      </p>
    )
  }
  // <output> tiene role="status" implícito: los lectores de pantalla anuncian el aviso
  return <output className={`${estilo} border-marca-200 border-l-marca-600 bg-marca-50 text-marca-900`}>{children}</output>
}

export function Boton({ cargando, disabled, children }) {
  return (
    <button
      type="submit"
      className="w-full rounded-md bg-marca-600 px-4 py-2.5 font-semibold text-white shadow-sm transition-colors hover:bg-marca-700 active:bg-marca-800 disabled:cursor-not-allowed disabled:bg-slate-400"
      disabled={cargando || disabled}
    >
      {cargando ? 'Procesando…' : children}
    </button>
  )
}

// Botones secundarios de las listas (Agregar, Editar, Eliminar, Cancelar)
export function BotonLinea({ children, onClick, peligro = false, disabled }) {
  const color = peligro ? 'border-red-300 text-red-700 hover:bg-red-50' : 'border-marca-700 text-marca-800 hover:bg-marca-50'
  return (
    <button type="button" onClick={onClick} disabled={disabled} className={`rounded-md border px-3 py-1.5 text-sm font-medium disabled:opacity-50 ${color}`}>
      {children}
    </button>
  )
}

const PASOS = [
  ['Regístrese', 'Valide su identidad con el Registro Civil y cree su contraseña.'],
  ['Complete el formulario', 'Datos de contacto, cargo, formación, experiencia y declaraciones.'],
  ['Envíe su postulación', 'Recibirá en su correo el formulario con su código único.'],
]

// Panel institucional que acompaña a los formularios de acceso (solo en pantallas anchas)
function PanelMarca() {
  return (
    <aside className="relative hidden overflow-hidden bg-marca-800 p-8 text-white lg:flex lg:flex-col lg:justify-between">
      <div className="absolute -right-24 -top-24 size-72 rounded-full bg-marca-700/60" aria-hidden="true" />
      <div className="absolute -bottom-32 -left-16 size-80 rounded-full bg-marca-900/50" aria-hidden="true" />
      <div className="relative">
        <p className="text-xs font-semibold uppercase tracking-widest text-acento-amarillo">Postulación en línea</p>
        <h2 className="mt-3 text-2xl font-semibold leading-snug">Miembros Principales y Alternos del Directorio del BIESS</h2>
        <p className="mt-3 text-sm text-marca-100">En representación de los Afiliados, Jubilados y Empleadores.</p>
      </div>
      <ol className="relative mt-10 flex flex-col gap-5">
        {PASOS.map(([titulo, texto], i) => (
          <li key={titulo} className="flex gap-4">
            <span className="flex size-8 shrink-0 items-center justify-center rounded-full border border-marca-300/60 text-sm font-semibold">{i + 1}</span>
            <span>
              <span className="block font-medium">{titulo}</span>
              <span className="block text-sm text-marca-100">{texto}</span>
            </span>
          </li>
        ))}
      </ol>
    </aside>
  )
}

export function Tarjeta({ titulo, children }) {
  return (
    <section className="mx-auto grid max-w-md overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm lg:max-w-5xl lg:grid-cols-[1fr_1.1fr]">
      <PanelMarca />
      <div className="p-6 sm:p-10">
        <h1 className="mb-6 text-2xl font-semibold text-marca-950">{titulo}</h1>
        {children}
      </div>
    </section>
  )
}
