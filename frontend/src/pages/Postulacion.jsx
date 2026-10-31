import { useState } from 'react'
import SelectorUbicacion from '../components/SelectorUbicacion.jsx'
import { UBICACION_VACIA } from '../services/catalogos.js'

// Fase 3 (#12) completa el formulario; por ahora solo la sección de domicilio (hito H1)
export default function Postulacion() {
  const [ubicacion, setUbicacion] = useState(UBICACION_VACIA)

  return (
    <section className="rounded-lg bg-white p-6 shadow-sm">
      <h1 className="mb-1 text-xl font-semibold">Formulario de postulación</h1>
      <h2 className="mb-4 text-slate-500">Información domiciliaria y de contacto</h2>
      <SelectorUbicacion value={ubicacion} onChange={setUbicacion} />
    </section>
  )
}
