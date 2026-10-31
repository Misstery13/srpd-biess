import { Component } from 'react'

// E4: un error de render muestra un aviso con reintento en vez de una pantalla en blanco
export default class ErrorBoundary extends Component {
  state = { error: null }

  static getDerivedStateFromError(error) {
    return { error }
  }

  componentDidCatch(error, info) {
    console.error(error, info.componentStack)
  }

  render() {
    if (!this.state.error) return this.props.children
    return (
      <section role="alert" className="mx-auto max-w-md rounded-lg bg-white p-6 text-center shadow-sm">
        <h1 className="text-xl font-semibold">Ocurrió un problema</h1>
        <p className="mt-2 text-slate-500">No se pudo mostrar esta pantalla.</p>
        <button
          type="button"
          onClick={() => this.setState({ error: null })}
          className="mt-4 rounded-md bg-emerald-700 px-4 py-2 font-medium text-white hover:bg-emerald-800"
        >
          Reintentar
        </button>
      </section>
    )
  }
}
