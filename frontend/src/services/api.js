import axios from 'axios'

const TOKEN_KEY = 'token'

// sessionStorage puede fallar (modo privado, bloqueado); la app sigue funcionando sin él
export function getToken() {
  try {
    return sessionStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function setToken(token) {
  try {
    if (token) sessionStorage.setItem(TOKEN_KEY, token)
    else sessionStorage.removeItem(TOKEN_KEY)
  } catch {
    // sin almacenamiento: el token vive solo en memoria del AuthContext (Fase 2)
  }
}

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  timeout: 15000,
})

api.interceptors.request.use((config) => {
  const token = getToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Todas las fallas llegan como Error con un mensaje listo para mostrar y el status HTTP
api.interceptors.response.use(
  (res) => res,
  (err) => {
    const status = err.response?.status
    const mensaje =
      err.response?.data?.error ||
      (status ? 'Ocurrió un error, inténtelo nuevamente.' : 'No se pudo conectar con el servidor.')
    return Promise.reject(Object.assign(new Error(mensaje), { status }))
  },
)

export default api
