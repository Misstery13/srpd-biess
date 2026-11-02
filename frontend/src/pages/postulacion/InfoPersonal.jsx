import { tienePareja } from '../../utils/ef12.js'
import { fechaUI } from './comun.js'
import Seccion from './Seccion.jsx'

// 1.1: datos del Registro Civil, solo lectura (P6); pareja si es casado/a o unión de hecho (1.1.7.1)
export default function InfoPersonal({ perfil }) {
  const datos = [
    ['Nombres', perfil.nombres],
    ['Apellidos', perfil.apellidos],
    ['Cédula', perfil.cedula],
    ['Género', perfil.genero],
    ['Fecha de nacimiento', fechaUI(perfil.fecha_nacimiento)],
    ['Nacionalidad', perfil.nacionalidad],
    ['Estado civil', perfil.estado_civil],
  ]
  if (tienePareja(perfil.estado_civil)) {
    datos.push(['Pareja: nombres y apellidos', perfil.conyuge_nombres], ['Pareja: cédula', perfil.conyuge_cedula])
  }

  return (
    <Seccion titulo="Información personal" descripcion="Datos obtenidos del Registro Civil del Ecuador. No se pueden modificar.">
      <dl className="grid gap-x-6 gap-y-3 sm:grid-cols-2" data-personal>
        {datos.map(([etiqueta, valor]) => (
          <div key={etiqueta}>
            <dt className="text-xs font-medium uppercase tracking-wide text-slate-500">{etiqueta}</dt>
            <dd className="text-slate-900">{valor || '—'}</dd>
          </div>
        ))}
      </dl>
    </Seccion>
  )
}
