import { useEffect, useMemo, useState } from 'react'
import { getToken, onNoAutorizado, setToken } from '../services/api.js'
import { AuthContext } from './auth.js'

// S2: el token vive en memoria (estado) y en sessionStorage para sobrevivir a un F5
export default function AuthProvider({ children }) {
  const [token, setTokenEstado] = useState(getToken)
  const [sesionExpirada, setSesionExpirada] = useState(false)

  useEffect(
    () =>
      onNoAutorizado(() => {
        setTokenEstado(null)
        setSesionExpirada(true)
      }),
    [],
  )

  const valor = useMemo(
    () => ({
      autenticado: !!token,
      sesionExpirada,
      iniciarSesion(nuevo) {
        setToken(nuevo)
        setTokenEstado(nuevo)
        setSesionExpirada(false)
      },
      cerrarSesion() {
        setToken(null)
        setTokenEstado(null)
      },
    }),
    [token, sesionExpirada],
  )

  return <AuthContext.Provider value={valor}>{children}</AuthContext.Provider>
}
