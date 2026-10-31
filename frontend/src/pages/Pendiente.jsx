export default function Pendiente({ titulo }) {
  return (
    <section className="mx-auto max-w-md rounded-lg bg-white p-6 shadow-sm">
      <h1 className="text-xl font-semibold">{titulo}</h1>
      <p className="mt-2 text-slate-500">Disponible en la Fase 2.</p>
    </section>
  )
}
