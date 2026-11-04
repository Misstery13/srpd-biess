import { Link, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../context/auth.js'
import ErrorBoundary from './ErrorBoundary.jsx'

export default function Layout() {
  const { pathname } = useLocation()
  const { autenticado, cerrarSesion } = useAuth()
  return (
    <div className="flex min-h-screen flex-col">
      <div className="franja-biess h-1.5 print:hidden" aria-hidden="true" />
      <header className="border-b border-slate-200 bg-white print:hidden">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3">
          <Link to="/" className="flex min-w-0 items-center gap-3 sm:gap-4">
            <img src="/logo-biess.png" alt="BIESS" width="480" height="175" className="h-9 w-auto shrink-0 sm:h-11" />
            <span className="hidden h-9 w-px bg-slate-200 sm:block" aria-hidden="true" />
            <span className="min-w-0">
              <span className="block truncate text-sm font-semibold leading-tight text-marca-900 sm:text-base">
                <span className="sm:hidden">Concurso de Méritos</span>
                <span className="hidden sm:inline">Concurso Público de Méritos y Oposición</span>
              </span>
              <span className="block truncate text-xs text-slate-500 sm:text-sm">
                Directorio del BIESS<span className="hidden sm:inline"> · Resolución C.D. 701</span>
              </span>
            </span>
          </Link>
          {autenticado && (
            <button
              type="button"
              onClick={() => cerrarSesion()}
              className="shrink-0 rounded-md border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:border-marca-600 hover:text-marca-700"
            >
              Cerrar sesión
            </button>
          )}
        </div>
      </header>
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-6 sm:py-10">
        {/* key: al navegar a otra ruta se limpia el error */}
        <ErrorBoundary key={pathname}>
          <Outlet />
        </ErrorBoundary>
      </main>
      <footer className="border-t border-slate-200 bg-white text-xs text-slate-500 print:hidden">
        <div className="mx-auto flex max-w-6xl flex-wrap justify-between gap-2 px-4 py-4">
          <span>Concurso Público de Méritos y Oposición 2026</span>
          <span>Sus datos se tratan conforme a la Ley Orgánica de Protección de Datos Personales.</span>
        </div>
      </footer>
    </div>
  )
}
