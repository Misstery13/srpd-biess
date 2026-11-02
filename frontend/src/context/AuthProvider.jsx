import { useEffect, useMemo, useState } from 'react'
import { getToken, onNoAutorizado, setToken as guardarToken } from '../services/api.js'
import { AuthContext } from './auth.js'

// S2: el token vive en memoria (estado) y en sessionStorage para sobrevivir a un F5
export default function AuthProvider({ children }) {
  const [token, setToken] = useState(getToken)
  const [aviso, setAviso] = useState('') // se muestra en /login al cerrarse la sesión

  useEffect(
    () =>
      onNoAutorizado(() => {
        setToken(null)
        setAviso('Su sesión terminó. Ingrese nuevamente.')
      }),
    [],
  )

  const valor = useMemo(
    () => ({
      autenticado: !!token,
      aviso,
      iniciarSesion(nuevo) {
        guardarToken(nuevo)
        setToken(nuevo)
        setAviso('')
      },
      cerrarSesion(mensaje = '') {
        guardarToken(null)
        setToken(null)
        setAviso(mensaje)
      },
    }),
    [token, aviso],
  )

  return <AuthContext.Provider value={valor}>{children}</AuthContext.Provider>
}
