import { Link, Navigate } from 'react-router-dom'
import { Tarjeta } from '../components/Formulario.jsx'
import { useAuth } from '../context/auth.js'

// EF-11 punto 3: opciones "Iniciar sesión" y "Registrarse"
export default function Inicio() {
  if (useAuth().autenticado) return <Navigate to="/postulacion" replace />
  return (
    <Tarjeta titulo="Bienvenido/a">
      <p className="text-slate-600">
        Ingrese para continuar con su postulación o regístrese si es la primera vez. Necesitará su cédula de identidad y su fecha de expedición o código dactilar.
      </p>
      <div className="mt-8 flex flex-col gap-3">
        <Link to="/login" className="rounded-md bg-marca-600 px-4 py-2.5 text-center font-semibold text-white shadow-sm transition-colors hover:bg-marca-700">
          Iniciar sesión
        </Link>
        <Link to="/registro" className="rounded-md border border-marca-600 px-4 py-2.5 text-center font-semibold text-marca-700 transition-colors hover:bg-marca-50">
          Registrarse
        </Link>
      </div>
    </Tarjeta>
  )
}
