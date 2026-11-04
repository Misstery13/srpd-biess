import { formatearFecha } from '../utils/validaciones.js'
import { Campo } from './Formulario.jsx'

// EF-11 punto 4: cédula + "fecha de expedición" (DD/MM/AAAA) ó "código dactilar", a elección del postulante
export default function CamposIdentidad({ value, onChange, errores, disabled }) {
  const cambiar = (campo, valor) => onChange({ ...value, [campo]: valor })

  return (
    <>
      <Campo
        id="cedula"
        etiqueta="Cédula de identidad"
        inputMode="numeric"
        autoComplete="username"
        maxLength={10}
        value={value.cedula}
        onChange={(e) => cambiar('cedula', e.target.value.replace(/\D/g, ''))}
        error={errores.cedula}
        disabled={disabled}
        required
      />

      <fieldset disabled={disabled}>
        <legend className="mb-1 text-sm font-medium">Verificar identidad con</legend>
        <div className="flex gap-4 text-sm">
          {[
            ['fecha', 'Fecha de expedición'],
            ['codigo', 'Código dactilar'],
          ].map(([metodo, texto]) => (
            <label key={metodo} className="flex items-center gap-2">
              <input type="radio" name="metodo" checked={value.metodo === metodo} onChange={() => cambiar('metodo', metodo)} />
              {texto}
            </label>
          ))}
        </div>
      </fieldset>

      {value.metodo === 'fecha' ? (
        <Campo
          id="fecha"
          etiqueta="Fecha de expedición (DD/MM/AAAA)"
          inputMode="numeric"
          placeholder="DD/MM/AAAA"
          maxLength={10}
          value={value.fecha}
          onChange={(e) => cambiar('fecha', formatearFecha(e.target.value))}
          error={errores.fecha}
          disabled={disabled}
          required
        />
      ) : (
        <Campo
          id="codigo"
          etiqueta="Código dactilar"
          placeholder="V4443V4444"
          maxLength={10}
          autoCapitalize="characters"
          value={value.codigo}
          onChange={(e) => cambiar('codigo', e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, ''))}
          error={errores.codigo}
          disabled={disabled}
          required
        />
      )}
    </>
  )
}
