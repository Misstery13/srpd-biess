import { useState } from 'react'
import { Campo } from '../../components/Formulario.jsx'
import SelectorUbicacion from '../../components/SelectorUbicacion.jsx'
import { UBICACION_VACIA } from '../../services/catalogos.js'
import { emailValido, esCelular, esNombre, esNumeroCasa, esTelefonoFijo, revisarTexto, soloTelefono } from '../../utils/validaciones.js'
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

function validar(f, emailRegistrado) {
  const e = {}
  const u = f.ubicacion
  const ecuador = !!u.provinciaId
  const [prov, ciudad] = [u.provinciaTexto.trim(), u.ciudadTexto.trim()]
  if (!u.paisId) e.ubicacion = 'Seleccione el país.'
  else if (ecuador && !u.cantonId) e.ubicacion = 'Seleccione la ciudad.'
  else if (!u.cantonId) {
    if (!prov || !ciudad) e.ubicacion = 'Indique la provincia y la ciudad.'
    else if (!esNombre(prov) || !esNombre(ciudad)) e.ubicacion = 'La provincia y la ciudad solo pueden tener letras.'
  }
  const cel = f.celular.trim()
  if (!cel) e.celular = 'Ingrese el teléfono celular.'
  else if (!esCelular(cel, ecuador)) e.celular = ecuador ? 'El celular tiene 10 dígitos y empieza con 09 (p. ej. 0991234567).' : 'Ingrese de 7 a 15 dígitos (puede empezar con +).'
  const tel = f.telefono.trim()
  if (tel && !esTelefonoFijo(tel, ecuador)) e.telefono = ecuador ? 'El teléfono fijo tiene 9 dígitos con el código de provincia (p. ej. 022345678).' : 'Ingrese de 7 a 15 dígitos (puede empezar con +).'
  const email2 = f.email2.trim()
  if (email2 && !emailValido(email2)) e.email2 = 'Ingrese un correo electrónico válido.'
  else if (email2 && email2.toLowerCase() === emailRegistrado?.toLowerCase()) e.email2 = 'Debe ser distinto del correo electrónico 1.'
  revisarTexto(e, 'calle_principal', f.calle_principal, { vacio: 'Ingrese la calle principal.' })
  if (f.numero.trim() && !esNumeroCasa(f.numero.trim())) e.numero = 'Use números y letras (p. ej. N34-120) o S/N.'
  revisarTexto(e, 'calle_secundaria', f.calle_secundaria)
  revisarTexto(e, 'sector', f.sector)
  revisarTexto(e, 'referencia', f.referencia)
  return e
}

// 1.2: información domiciliaria y de contacto
export default function Domicilio({ postulacion, perfil, guardar, soloLectura }) {
  const [form, setForm] = useState(() => desdeApi(postulacion?.domicilio))
  const [errores, setErrores] = useState({})
  const campo = (nombre, filtro = (v) => v) => ({ value: form[nombre], onChange: (e) => setForm({ ...form, [nombre]: filtro(e.target.value) }), error: errores[nombre] })

  async function onGuardar() {
    const e = validar(form, perfil.email)
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
        <Campo id="celular" etiqueta="Teléfono celular" inputMode="tel" maxLength={15} required {...campo('celular', soloTelefono)} />
        <Campo id="telefono" etiqueta="Teléfono domicilio" inputMode="tel" maxLength={15} {...campo('telefono', soloTelefono)} />
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
