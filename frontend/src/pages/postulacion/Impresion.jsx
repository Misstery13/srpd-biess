import { CONSENTIMIENTOS, INHABILIDADES, TIPOS_EVENTO, TIPOS_EXPERIENCIA, TIPOS_FORMACION, TIPOS_PUBLICACION, calcularDuracion, cargoPorCodigo, deFechaApi, textoDuracion, tienePareja } from '../../utils/ef12.js'
import { fechaUI, nombreOpcion } from './comun.js'

const siNo = (v) => (v == null ? '—' : v ? 'Si' : 'No')

function Bloque({ titulo, filas }) {
  return (
    <section className="mb-4 break-inside-avoid">
      <h2 className="mb-1 border-b border-slate-400 text-sm font-bold uppercase">{titulo}</h2>
      <table className="w-full text-xs">
        <tbody>
          {filas.map(([k, v], i) => (
            <tr key={i} className="align-top">
              <th className="w-1/3 py-0.5 pr-2 text-left font-medium">{k}</th>
              <td className="py-0.5">{v || '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}

// Formulario final para imprimir y firmar (1.10.1 punto 5 y 1.10.2). Solo se ve al imprimir.
export default function Impresion({ perfil, postulacion: p }) {
  const d = p?.domicilio
  const cargo = cargoPorCodigo(p?.cargo)
  const lista = (items, fila) => (items?.length ? items.flatMap(fila) : [['Registros', 'Ninguno']])

  return (
    <article className="hidden text-black print:block">
      <header className="mb-4 text-center">
        <h1 className="text-base font-bold">FORMULARIO DE POSTULACIÓN</h1>
        <p className="text-sm">Concurso Público de Méritos y Oposición — Directorio del BIESS</p>
        <p className="text-sm font-semibold">Código: {p?.codigo ?? 'sin guardar'}</p>
      </header>

      <Bloque
        titulo="Información personal"
        filas={[
          ['Nombres', perfil.nombres],
          ['Apellidos', perfil.apellidos],
          ['Cédula', perfil.cedula],
          ['Género', perfil.genero],
          ['Fecha de nacimiento', fechaUI(perfil.fecha_nacimiento)],
          ['Nacionalidad', perfil.nacionalidad],
          ['Estado civil', perfil.estado_civil],
          ...(tienePareja(perfil.estado_civil) ? [['Pareja', `${perfil.conyuge_nombres ?? ''} — ${perfil.conyuge_cedula ?? ''}`]] : []),
        ]}
      />
      <Bloque
        titulo="Información domiciliaria y de contacto"
        filas={[
          ['Teléfono celular', d?.celular],
          ['Correo electrónico 1', perfil.email],
          ['Correo electrónico 2', d?.email2],
          ['Dirección', d && [d.calle_principal, d.numero, d.calle_secundaria].filter(Boolean).join(' ')],
          ['Sector / Referencia', d && [d.sector, d.referencia].filter(Boolean).join(' — ')],
          ['Provincia / Ciudad (texto)', d && [d.provincia_texto, d.ciudad_texto].filter(Boolean).join(' / ')],
          ['Teléfono domicilio', d?.telefono],
        ]}
      />
      <Bloque titulo="Información adicional — familiares en IESS/BIESS" filas={lista(p?.familiares, (f) => [[f.nombres, `${f.parentesco} · ${f.institucion} · ${f.area ?? ''} · ${f.cargo ?? ''}`]])} />
      <Bloque
        titulo="Cargo al que postula"
        filas={[
          ['Cargo', cargo?.nombre],
          ['Respaldo de organización', siNo(p?.respaldo_org)],
          ['Nombre de organización', p?.org_nombre],
        ]}
      />
      <Bloque titulo="Formación académica" filas={lista(p?.formacion, (f) => [[f.titulo, `${f.nivel} · ${f.institucion} (${nombreOpcion(TIPOS_FORMACION, f.tipo)}) · ${f.area} · SENESCYT ${f.senescyt_numero ?? '—'} ${fechaUI(f.senescyt_fecha)}`]])} />
      <Bloque
        titulo="Experiencia profesional"
        filas={lista(p?.experiencia, (e) => [
          [e.cargo, `${nombreOpcion(TIPOS_EXPERIENCIA, e.tipo)} · ${e.institucion} · ${e.area} · ${fechaUI(e.fecha_inicio)} – ${e.fecha_fin ? fechaUI(e.fecha_fin) : 'actualidad'} · ${textoDuracion(calcularDuracion(deFechaApi(e.fecha_inicio), deFechaApi(e.fecha_fin)))}`],
        ])}
      />
      <Bloque titulo="Publicaciones" filas={lista(p?.publicaciones, (x) => [[x.titulo, `${nombreOpcion(TIPOS_PUBLICACION, x.tipo)} · ${x.medio} · ${fechaUI(x.fecha)} · Relación profesional: ${siNo(x.relacion_profesional)}`]])} />
      <Bloque titulo="Capacitación" filas={lista(p?.capacitaciones, (c) => [[c.nombre_evento, `${nombreOpcion(TIPOS_EVENTO, c.tipo_evento)} · ${c.institucion} · ${fechaUI(c.fecha_inicio)} – ${fechaUI(c.fecha_fin)} · ${c.horas} h`]])} />
      <Bloque
        titulo="Inhabilidades"
        filas={[
          ...INHABILIDADES.map(([clave, pregunta], i) => [`${i + 1}. ${pregunta}`, siNo(p?.inhabilidades?.[clave])]),
          ...(p?.inhabilidades?.removido_control ? [['Removido: institución / cargo / fecha / motivo', [p.inhabilidades.removido_institucion, p.inhabilidades.removido_cargo, fechaUI(p.inhabilidades.removido_fecha_fin), p.inhabilidades.removido_motivo].filter(Boolean).join(' · ')]] : []),
        ]}
      />

      <section className="mb-6 break-inside-avoid text-xs">
        {CONSENTIMIENTOS.map((t, i) => (
          <p key={i} className="mb-1">
            {i + 1}. {t}
          </p>
        ))}
      </section>

      {/* 1.10.2: firma, nombres y apellidos y número de cédula */}
      <footer className="mt-16 flex break-inside-avoid flex-col items-center text-sm">
        <div className="w-72 border-t border-black" />
        <p>Firma del postulante</p>
        <p className="font-semibold">
          {perfil.nombres} {perfil.apellidos}
        </p>
        <p>C.I. {perfil.cedula}</p>
      </footer>
    </article>
  )
}
