import { useEffect, useMemo, useState } from 'react'
import { getToken, onNoAutorizado, setToken as guardarToken } from '../services/api.js'
import { AuthContext } from './auth.js'

// S2: el token vive en memoria (estado) y en sessionStorage para sobrevivir a un F5
export default function AuthProvider({ children }) {
  const [token, setToken] = useState(getToken)
  const [sesionExpirada, setSesionExpirada] = useState(false)

  useEffect(
    () =>
      onNoAutorizado(() => {
        setToken(null)
        setSesionExpirada(true)
      }),
    [],
  )

  const valor = useMemo(
    () => ({
      autenticado: !!token,
      sesionExpirada,
      iniciarSesion(nuevo) {
        guardarToken(nuevo)
        setToken(nuevo)
        setSesionExpirada(false)
      },
      cerrarSesion() {
        guardarToken(null)
        setToken(null)
      },
    }),
    [token, sesionExpirada],
  )

  return <AuthContext.Provider value={valor}>{children}</AuthContext.Provider>
}
