import { useState } from 'react'
import { subirDocumento } from '../services/postulacion.js'
import { MB } from '../utils/ef12.js'

// Documento de respaldo: se valida en el navegador (formato y tamaño, A3) y se sube al elegirlo.
// El backend vuelve a validar el tipo real (%PDF, A2) y guarda el archivo con nombre UUID (A1).
export default function CampoPDF({ id, etiqueta, value, onChange, limite, tipo, obligatorio, error }) {
  const [estado, setEstado] = useState({ subiendo: false, error: '', nombre: '' })

  async function elegir(e) {
    const archivo = e.target.files?.[0]
    e.target.value = '' // permite volver a elegir el mismo archivo
    if (!archivo) return
    const esPdf = archivo.type === 'application/pdf' || archivo.name.toLowerCase().endsWith('.pdf')
    if (!esPdf) return setEstado({ subiendo: false, error: 'El documento debe estar en formato PDF.', nombre: '' })
    if (archivo.size > limite) return setEstado({ subiendo: false, error: `El documento supera el máximo de ${limite / MB} MB.`, nombre: '' })

    setEstado({ subiendo: true, error: '', nombre: archivo.name })
    try {
      onChange(await subirDocumento(archivo, tipo))
      setEstado({ subiendo: false, error: '', nombre: archivo.name })
    } catch (err) {
      setEstado({ subiendo: false, error: err.message, nombre: '' })
    }
  }

  const mensaje = estado.error || error
  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium">
        {etiqueta} (PDF, máximo {limite / MB} MB)
        {obligatorio && <span className="text-red-600"> *</span>}
      </label>
      <input
        id={id}
        type="file"
        accept="application/pdf,.pdf"
        onChange={elegir}
        aria-invalid={mensaje ? true : undefined}
        aria-describedby={mensaje ? `${id}-error` : undefined}
        className="block w-full text-sm file:mr-3 file:rounded-md file:border-0 file:bg-emerald-50 file:px-3 file:py-2 file:text-emerald-800"
      />
      {estado.subiendo && <p className="mt-1 text-sm text-slate-500">Subiendo {estado.nombre}…</p>}
      {!estado.subiendo && value && (
        <p className="mt-1 text-sm text-emerald-800" data-documento>
          ✓ Documento cargado{estado.nombre && `: ${estado.nombre}`}
        </p>
      )}
      {mensaje && (
        <p id={`${id}-error`} role="alert" className="mt-1 text-sm text-red-700">
          {mensaje}
        </p>
      )}
    </div>
  )
}
