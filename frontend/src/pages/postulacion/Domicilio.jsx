import { useState } from 'react'
import { Campo } from '../../components/Formulario.jsx'
import SelectorUbicacion from '../../components/SelectorUbicacion.jsx'
import { UBICACION_VACIA } from '../../services/catalogos.js'
import { emailValido } from '../../utils/validaciones.js'
import Seccion from './Seccion.jsx'

const texto = (v) => (v?.trim() ? v.trim() : null)
const numero = (v) => (v ? Number(v) : null)
const cadena = (v) => (v == null ? '' : String(v))

// API (domicilios) ↔ formulario
function desdeApi(d) {
  if (!d) return { ubicacion: UBICACION_VACIA, celular: '', email2: '', calle_principal: '', numero: '', calle_secundaria: '', sector: '', referencia: '', telefono: '' }
  return {
    ubicacion: {
      paisId: cadena(d.pais_id),
      provinciaId: cadena(d.provincia_id),
      cantonId: cadena(d.canton_id),
      provinciaTexto: d.provincia_texto ?? '',
      ciudadTexto: d.ciudad_texto ?? '',
    },
    celular: d.celular ?? '',
    email2: d.email2 ?? '',
    calle_principal: d.calle_principal ?? '',
    numero: d.numero ?? '',
    calle_secundaria: d.calle_secundaria ?? '',
    sector: d.sector ?? '',
    referencia: d.referencia ?? '',
    telefono: d.telefono ?? '',
  }
}

function aApi(f) {
  const u = f.ubicacion
  const extranjero = !u.provinciaId && !u.cantonId
  return {
    pais_id: numero(u.paisId),
    provincia_id: numero(u.provinciaId),
    canton_id: numero(u.cantonId),
    provincia_texto: extranjero ? texto(u.provinciaTexto) : null,
    ciudad_texto: extranjero ? texto(u.ciudadTexto) : null,
    celular: f.celular.trim(),
    email2: texto(f.email2),
    calle_principal: f.calle_principal.trim(),
    numero: texto(f.numero),
    calle_secundaria: texto(f.calle_secundaria),
    sector: texto(f.sector),
    referencia: texto(f.referencia),
    telefono: texto(f.telefono),
  }
}

const telefonoValido = (t) => /^\+?\d{7,15}$/.test(t)

function validar(f) {
  const e = {}
  const u = f.ubicacion
  if (!u.paisId) e.ubicacion = 'Seleccione el país.'
  else if (!u.cantonId && !u.ciudadTexto.trim()) e.ubicacion = 'Indique la provincia y la ciudad.'
  if (!telefonoValido(f.celular.trim())) e.celular = 'Ingrese un número de celular válido (p. ej. 0991234567).'
  if (f.email2.trim() && !emailValido(f.email2.trim())) e.email2 = 'Ingrese un correo electrónico válido.'
  if (!f.calle_principal.trim()) e.calle_principal = 'Ingrese la calle principal.'
  if (f.telefono.trim() && !telefonoValido(f.telefono.trim())) e.telefono = 'Ingrese solo números (7 a 15 dígitos).'
  return e
}

// 1.2: información domiciliaria y de contacto
export default function Domicilio({ postulacion, perfil, guardar, soloLectura }) {
  const [form, setForm] = useState(() => desdeApi(postulacion?.domicilio))
  const [errores, setErrores] = useState({})
  const campo = (nombre) => ({ value: form[nombre], onChange: (e) => setForm({ ...form, [nombre]: e.target.value }), error: errores[nombre] })

  async function onGuardar() {
    const e = validar(form)
    setErrores(e)
    if (Object.keys(e).length) return e
    await guardar({ domicilio: aApi(form) })
  }

  return (
    <Seccion titulo="Información domiciliaria y de contacto" onGuardar={onGuardar} soloLectura={soloLectura}>
      <div>
        <SelectorUbicacion value={form.ubicacion} onChange={(ubicacion) => setForm({ ...form, ubicacion })} />
        {errores.ubicacion && (
          <p role="alert" className="mt-1 text-sm text-red-700">
            {errores.ubicacion}
          </p>
        )}
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <Campo id="celular" etiqueta="Teléfono celular" inputMode="tel" maxLength={15} required {...campo('celular')} />
        <Campo id="telefono" etiqueta="Teléfono domicilio" inputMode="tel" maxLength={15} {...campo('telefono')} />
        <Campo id="email1" etiqueta="Correo electrónico 1" value={perfil.email} readOnly disabled ayuda="Correo registrado en su cuenta." />
        <Campo id="email2" etiqueta="Correo electrónico 2" type="email" maxLength={150} {...campo('email2')} />
        <Campo id="calle_principal" etiqueta="Calle principal domicilio" maxLength={150} required {...campo('calle_principal')} />
        <Campo id="numero" etiqueta="Número" maxLength={20} {...campo('numero')} />
        <Campo id="calle_secundaria" etiqueta="Calle secundaria" maxLength={150} {...campo('calle_secundaria')} />
        <Campo id="sector" etiqueta="Sector" maxLength={100} {...campo('sector')} />
      </div>
      <Campo id="referencia" etiqueta="Referencia" maxLength={255} {...campo('referencia')} />
    </Seccion>
  )
}
