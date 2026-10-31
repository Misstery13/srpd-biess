import { useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import CamposIdentidad from '../components/CamposIdentidad.jsx'
import { Alerta, Boton, Campo, Tarjeta } from '../components/Formulario.jsx'
import { useAuth } from '../context/auth.js'
import useBloqueo from '../hooks/useBloqueo.js'
import { MENSAJES, registrar, verificar } from '../services/auth.js'
import { emailValido, IDENTIDAD_VACIA, validarIdentidad, validarPassword } from '../utils/validaciones.js'

// EF-11 punto 4: (1) verificar identidad con el Registro Civil → (2) correo y contraseña
export default function Registro() {
  const { autenticado } = useAuth()
  const navigate = useNavigate()
  const [identidad, setIdentidad] = useState(IDENTIDAD_VACIA)
  const [nombres, setNombres] = useState(null) // != null → paso 2
  const [cuenta, setCuenta] = useState({ email: '', password: '', confirmar: '' })
  const [errores, setErrores] = useState({})
  const [error, setError] = useState('')
  const [cargando, setCargando] = useState(false)
  const [bloqueado, bloquear] = useBloqueo()

  if (autenticado) return <Navigate to="/postulacion" replace />

  async function ejecutar(errs, accion) {
    setErrores(errs)
    setError('')
    if (Object.keys(errs).length) return
    setCargando(true)
    try {
      await accion()
    } catch (err) {
      if (err.status === 429) bloquear()
      setError(err.status === 401 ? MENSAJES.identidadIncorrecta : err.status === 429 ? MENSAJES.registroBloqueado : err.message)
    }
    setCargando(false)
  }

  function enviarIdentidad(e) {
    e.preventDefault()
    const { errores: errs, datos } = validarIdentidad(identidad)
    ejecutar(errs, async () => setNombres((await verificar(datos)).nombres))
  }

  function enviarCuenta(e) {
    e.preventDefault()
    const errs = validarPassword(cuenta.password, cuenta.confirmar)
    if (!emailValido(cuenta.email)) errs.email = 'Ingrese un correo electrónico válido.'
    ejecutar(errs, async () => {
      await registrar({ cedula: identidad.cedula, email: cuenta.email.trim(), password: cuenta.password })
      // EF-11 punto 5: registro correcto → pantalla principal para ingresar con cédula y contraseña
      navigate('/login', { replace: true, state: { aviso: 'Registro exitoso. Ingrese con su cédula y contraseña.' } })
    })
  }

  if (nombres === null) {
    return (
      <Tarjeta titulo="Registrarse">
        <p className="mb-4 text-sm text-slate-500">Paso 1 de 2 · Verificación de identidad con el Registro Civil</p>
        <form onSubmit={enviarIdentidad} noValidate className="flex flex-col gap-4">
          <Alerta>{error}</Alerta>
          <CamposIdentidad value={identidad} onChange={setIdentidad} errores={errores} disabled={bloqueado} />
          <Boton cargando={cargando} disabled={bloqueado}>
            Verificar
          </Boton>
        </form>
        <p className="mt-4 text-sm">
          ¿Ya tiene cuenta?{' '}
          <Link to="/login" className="text-emerald-800 underline">
            Iniciar sesión
          </Link>
        </p>
      </Tarjeta>
    )
  }

  const cambiar = (campo) => (e) => setCuenta({ ...cuenta, [campo]: e.target.value })
  return (
    <Tarjeta titulo="Registrarse">
      <p className="mb-4 text-sm text-slate-500">Paso 2 de 2 · Datos de acceso</p>
      <form onSubmit={enviarCuenta} noValidate className="flex flex-col gap-4">
        <Alerta tipo="info">Identidad verificada: {nombres} ({identidad.cedula})</Alerta>
        <Alerta>{error}</Alerta>
        <Campo id="email" etiqueta="Correo electrónico" type="email" autoComplete="email" maxLength={150} value={cuenta.email} onChange={cambiar('email')} error={errores.email} required />
        <Campo id="password" etiqueta="Contraseña" type="password" autoComplete="new-password" value={cuenta.password} onChange={cambiar('password')} error={errores.password} required />
        <Campo id="confirmar" etiqueta="Confirmar contraseña" type="password" autoComplete="new-password" value={cuenta.confirmar} onChange={cambiar('confirmar')} error={errores.confirmar} required />
        <p className="-mt-2 text-xs text-slate-500">Mínimo 8 caracteres, con al menos una letra y un número.</p>
        <Boton cargando={cargando}>Registrarme</Boton>
      </form>
    </Tarjeta>
  )
}
