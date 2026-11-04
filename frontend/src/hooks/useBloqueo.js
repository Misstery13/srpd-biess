import { useCallback, useEffect, useRef, useState } from 'react'

// S4: tras un 429 el botón queda deshabilitado 10 min. El backend es quien realmente bloquea;
// si se recarga la página, el siguiente intento vuelve a recibir 429 y se bloquea de nuevo.
export default function useBloqueo(ms = 10 * 60 * 1000) {
  const [bloqueado, setBloqueado] = useState(false)
  const timer = useRef(null)

  useEffect(() => () => clearTimeout(timer.current), [])

  const bloquear = useCallback(() => {
    setBloqueado(true)
    clearTimeout(timer.current)
    timer.current = setTimeout(() => setBloqueado(false), ms)
  }, [ms])

  return [bloqueado, bloquear]
}
