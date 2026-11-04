// Utilidades compartidas por las secciones del formulario
export const hoy = () => {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// Texto visible de una opción [['CODIGO', 'Texto'], ...]
export const nombreOpcion = (opciones, codigo) => opciones.find(([c]) => c === codigo)?.[1] ?? codigo

export const fechaUI = (iso) => (iso ? iso.slice(0, 10).split('-').reverse().join('/') : '')

// Lo mínimo que el backend exige antes de enviar (EnviarPostulacion)
export function faltantes(p) {
  const f = []
  if (!p?.domicilio) f.push('Información domiciliaria y de contacto')
  if (!p?.cargo) f.push('Selección del cargo al que postula')
  if (!p?.inhabilidades) f.push('Inhabilidades')
  return f
}
