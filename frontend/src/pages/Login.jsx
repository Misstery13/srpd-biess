import { useState } from 'react'
import { Link, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { Alerta, Boton, Campo, Tarjeta } from '../components/Formulario.jsx'
import { useAuth } from '../context/auth.js'
import useBloqueo from '../hooks/useBloqueo.js'
import { login, MENSAJES } from '../services/auth.js'
import { cedulaValida } from '../utils/validaciones.js'

// EF-11 puntos 5 y 6: ingreso con cédula y contraseña
export default function Login() {
  const { autenticado, sesionExpirada, iniciarSesion } = useAuth()
  const aviso = useLocation().state?.aviso
  const navigate = useNavigate()
  const [form, setForm] = useState({ cedula: '', password: '' })
  const [errores, setErrores] = useState({})
  const [error, setError] = useState('')
  const [cargando, setCargando] = useState(false)
  const [bloqueado, bloquear] = useBloqueo()

  if (autenticado) return <Navigate to="/postulacion" replace />

  async function enviar(e) {
    e.preventDefault()
    const errs = {}
    if (!cedulaValida(form.cedula)) errs.cedula = 'Ingrese una cédula ecuatoriana válida.'
    if (!form.password) errs.password = 'Ingrese su contraseña.'
    setErrores(errs)
    setError('')
    if (Object.keys(errs).length) return

    setCargando(true)
    try {
      const { token } = await login(form)
      iniciarSesion(token)
      // EF-12: al ingresar se dirige obligatoriamente al formulario de postulación
      navigate('/postulacion', { replace: true })
    } catch (err) {
      if (err.status === 429) bloquear()
      setError(err.status === 401 ? MENSAJES.credencialesIncorrectas : err.status === 429 ? MENSAJES.loginBloqueado : err.message)
      setCargando(false)
    }
  }

  return (
    <Tarjeta titulo="Iniciar sesión">
      <form onSubmit={enviar} noValidate className="flex flex-col gap-4">
        {sesionExpirada && <Alerta tipo="info">Su sesión terminó. Ingrese nuevamente.</Alerta>}
        <Alerta tipo="info">{aviso}</Alerta>
        <Alerta>{error}</Alerta>
        <Campo
          id="cedula"
          etiqueta="Cédula de identidad"
          inputMode="numeric"
          autoComplete="username"
          maxLength={10}
          value={form.cedula}
          onChange={(e) => setForm({ ...form, cedula: e.target.value.replace(/\D/g, '') })}
          error={errores.cedula}
          required
        />
        <Campo
          id="password"
          etiqueta="Contraseña"
          type="password"
          autoComplete="current-password"
          value={form.password}
          onChange={(e) => setForm({ ...form, password: e.target.value })}
          error={errores.password}
          required
        />
        <Boton cargando={cargando} disabled={bloqueado}>
          Ingresar
        </Boton>
      </form>
      <div className="mt-4 flex justify-between text-sm">
        <Link to="/recuperar" className="text-emerald-800 underline">
          ¿Olvidaste tu contraseña?
        </Link>
        <Link to="/registro" className="text-emerald-800 underline">
          Registrarse
        </Link>
      </div>
    </Tarjeta>
  )
}
