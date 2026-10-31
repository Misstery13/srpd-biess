import { useState } from 'react'
import { Link } from 'react-router-dom'
import CamposIdentidad from '../components/CamposIdentidad.jsx'
import { Alerta, Boton, Tarjeta } from '../components/Formulario.jsx'
import useBloqueo from '../hooks/useBloqueo.js'
import { mensajeError, MENSAJES, recuperar } from '../services/auth.js'
import { IDENTIDAD_VACIA, validarIdentidad } from '../utils/validaciones.js'

// EF-11 punto 6: cédula + fecha de expedición ó código dactilar → correo para restablecer la contraseña
export default function Recuperar() {
  const [identidad, setIdentidad] = useState(IDENTIDAD_VACIA)
  const [errores, setErrores] = useState({})
  const [error, setError] = useState('')
  const [enviado, setEnviado] = useState(false)
  const [cargando, setCargando] = useState(false)
  const [bloqueado, bloquear] = useBloqueo()

  async function enviar(e) {
    e.preventDefault()
    const { errores: errs, datos } = validarIdentidad(identidad)
    setErrores(errs)
    setError('')
    if (Object.keys(errs).length) return
    setCargando(true)
    try {
      await recuperar(datos)
      setEnviado(true)
    } catch (err) {
      if (err.status === 429) bloquear()
      setError(mensajeError(err, MENSAJES.identidadIncorrecta))
    }
    setCargando(false)
  }

  return (
    <Tarjeta titulo="Recuperar contraseña">
      {enviado ? (
        <Alerta tipo="info">
          Le enviamos un correo electrónico con el enlace para restablecer su contraseña. El enlace vence en 15 minutos.
        </Alerta>
      ) : (
        <form onSubmit={enviar} noValidate className="flex flex-col gap-4">
          <Alerta>{error}</Alerta>
          <CamposIdentidad value={identidad} onChange={setIdentidad} errores={errores} disabled={bloqueado} />
          <Boton cargando={cargando} disabled={bloqueado}>
            Enviar correo
          </Boton>
        </form>
      )}
      <Link to="/login" className="mt-4 inline-block text-sm text-emerald-800 underline">
        Volver a iniciar sesión
      </Link>
    </Tarjeta>
  )
}
