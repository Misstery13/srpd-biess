import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { Alerta, Boton, Campo, Tarjeta } from '../components/Formulario.jsx'
import { restablecer } from '../services/auth.js'
import { validarPassword } from '../utils/validaciones.js'

// Destino del enlace del correo de recuperación: /restablecer?token=...
export default function Restablecer() {
  const token = useSearchParams()[0].get('token')
  const navigate = useNavigate()
  const [form, setForm] = useState({ password: '', confirmar: '' })
  const [errores, setErrores] = useState({})
  const [error, setError] = useState(token ? '' : 'El enlace de recuperación no es válido.')
  const [cargando, setCargando] = useState(false)

  async function enviar(e) {
    e.preventDefault()
    const errs = validarPassword(form.password, form.confirmar)
    setErrores(errs)
    setError('')
    if (Object.keys(errs).length) return
    setCargando(true)
    try {
      await restablecer({ token, password: form.password })
      navigate('/login', { replace: true, state: { aviso: 'Su contraseña fue restablecida. Ingrese con su cédula y la nueva contraseña.' } })
    } catch (err) {
      setError(err.message)
      setCargando(false)
    }
  }

  const cambiar = (campo) => (e) => setForm({ ...form, [campo]: e.target.value })
  return (
    <Tarjeta titulo="Restablecer contraseña">
      <form onSubmit={enviar} noValidate className="flex flex-col gap-4">
        <Alerta>{error}</Alerta>
        <Campo id="password" etiqueta="Nueva contraseña" type="password" autoComplete="new-password" value={form.password} onChange={cambiar('password')} error={errores.password} disabled={!token} required />
        <Campo id="confirmar" etiqueta="Confirmar contraseña" type="password" autoComplete="new-password" value={form.confirmar} onChange={cambiar('confirmar')} error={errores.confirmar} disabled={!token} required />
        <p className="-mt-2 text-xs text-slate-500">Mínimo 8 caracteres, con al menos una letra y un número.</p>
        <Boton cargando={cargando} disabled={!token}>
          Guardar contraseña
        </Boton>
      </form>
      <Link to="/recuperar" className="mt-4 inline-block text-sm text-emerald-800 underline">
        Solicitar un nuevo enlace
      </Link>
    </Tarjeta>
  )
}
