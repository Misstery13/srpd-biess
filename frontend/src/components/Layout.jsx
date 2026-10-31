import { Link, Outlet, useLocation } from 'react-router-dom'
import ErrorBoundary from './ErrorBoundary.jsx'

export default function Layout() {
  const { pathname } = useLocation()
  return (
    <div className="min-h-screen bg-slate-50 text-slate-800">
      <header className="bg-emerald-800 text-white">
        <div className="mx-auto max-w-4xl px-4 py-4">
          <Link to="/" className="block text-lg font-semibold leading-tight">
            Concurso Público de Méritos y Oposición
          </Link>
          <p className="text-sm text-emerald-100">Directorio del BIESS</p>
        </div>
      </header>
      <main className="mx-auto max-w-4xl px-4 py-8">
        {/* key: al navegar a otra ruta se limpia el error */}
        <ErrorBoundary key={pathname}>
          <Outlet />
        </ErrorBoundary>
      </main>
    </div>
  )
}
