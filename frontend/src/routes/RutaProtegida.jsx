import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../context/auth.js'

// Sin sesión → /login (también cuando un 401 la cierra, E3)
export default function RutaProtegida() {
  return useAuth().autenticado ? <Outlet /> : <Navigate to="/login" replace />
}
