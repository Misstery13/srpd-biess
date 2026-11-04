import api from './api.js'

export const UBICACION_VACIA = { paisId: '', provinciaId: '', cantonId: '', provinciaTexto: '', ciudadTexto: '' }

// VITE_MOCK=true usa datos locales con la misma forma que el contrato de API
const MOCK = import.meta.env.VITE_MOCK === 'true'

const mock = {
  paises: [
    { id: 42, nombre: 'Colombia' },
    { id: 52, nombre: 'Ecuador' },
    { id: 59, nombre: 'España' },
    { id: 142, nombre: 'Perú' },
  ],
  provincias: {
    52: [
      { id: 1, nombre: 'Azuay' },
      { id: 10, nombre: 'Guayas' },
      { id: 19, nombre: 'Pichincha' },
    ],
  },
  cantones: {
    1: [{ id: 1, nombre: 'Cuenca' }, { id: 2, nombre: 'Gualaceo' }],
    10: [{ id: 69, nombre: 'Guayaquil' }, { id: 70, nombre: 'Durán' }, { id: 72, nombre: 'Samborondón' }],
    19: [{ id: 160, nombre: 'Quito' }, { id: 161, nombre: 'Cayambe' }, { id: 166, nombre: 'Rumiñahui' }],
  },
}

const demora = (data) => new Promise((resolve) => setTimeout(() => resolve(data), 300))

export async function getPaises() {
  if (MOCK) return demora(mock.paises)
  return (await api.get('/catalogos/paises')).data
}

export async function getProvincias(paisId) {
  if (MOCK) return demora(mock.provincias[paisId] ?? [])
  return (await api.get(`/catalogos/paises/${paisId}/provincias`)).data
}

export async function getCantones(provinciaId) {
  if (MOCK) return demora(mock.cantones[provinciaId] ?? [])
  return (await api.get(`/catalogos/provincias/${provinciaId}/cantones`)).data
}
