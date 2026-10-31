import { Link } from 'react-router-dom'

export default function Inicio() {
  return (
    <section className="mx-auto max-w-md rounded-lg bg-white p-6 shadow-sm">
      <h1 className="mb-6 text-xl font-semibold">Bienvenido/a</h1>
      <div className="flex flex-col gap-3 sm:flex-row">
        <Link to="/login" className="flex-1 rounded-md bg-emerald-700 px-4 py-2 text-center font-medium text-white hover:bg-emerald-800">
          Iniciar sesión
        </Link>
        <Link to="/registro" className="flex-1 rounded-md border border-emerald-700 px-4 py-2 text-center font-medium text-emerald-800 hover:bg-emerald-50">
          Registrarse
        </Link>
      </div>
    </section>
  )
}
