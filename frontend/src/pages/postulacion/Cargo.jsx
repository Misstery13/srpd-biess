import { useEffect, useState } from 'react'
import CampoPDF from '../../components/CampoPDF.jsx'
import { Alerta, Campo, SiNo } from '../../components/Formulario.jsx'
import { validarCargo } from '../../services/postulacion.js'
import { CARGOS, cargoPorCodigo, LIMITE_ORGANIZACION } from '../../utils/ef12.js'
import Seccion from './Seccion.jsx'

// 1.4: selección del cargo. La calidad de afiliado / jubilado / empleador se valida contra el
// Historial Laboral del IESS; si no la cumple, se muestra el mensaje del documento y la postulación se cierra.
export default function Cargo({ postulacion, guardar, soloLectura, onCierre }) {
  const [cargo, setCargo] = useState(postulacion?.cargo ?? '')
  const [validacion, setValidacion] = useState({ cargo: '', habilitado: null, error: '' }) // resultado para `cargo`
  const [respaldo, setRespaldo] = useState(postulacion?.respaldo_org ?? null)
  const [orgNombre, setOrgNombre] = useState(postulacion?.org_nombre ?? '')
  const [orgDocumento, setOrgDocumento] = useState(postulacion?.org_documento ?? null)
  const [errores, setErrores] = useState({})

  // Valida el cargo elegido (y el guardado, al abrir la sección)
  useEffect(() => {
    if (!cargo) return
    let vigente = true
    validarCargo(cargo)
      .then(({ habilitado }) => {
        if (!vigente) return
        setValidacion({ cargo, habilitado, error: '' })
        const c = cargoPorCodigo(cargo)
        if (!habilitado) onCierre(`${c.noHabilitado}. ${c.cierre}`)
      })
      .catch((err) => vigente && setValidacion({ cargo, habilitado: null, error: err.message }))
    return () => {
      vigente = false
    }
  }, [cargo, onCierre])

  const datos = cargoPorCodigo(cargo)
  const validando = cargo && validacion.cargo !== cargo
  const habilitado = !validando && validacion.habilitado === true

  async function onGuardar() {
    const e = {}
    if (!cargo) e.cargo = 'Seleccione el cargo al que postula.'
    else if (!habilitado) e.cargo = 'Espere la validación del cargo.'
    if (habilitado && respaldo === null) e.respaldo = 'Responda si cuenta con el respaldo de una organización.'
    if (respaldo) {
      if (!orgNombre.trim()) e.orgNombre = 'Ingrese el nombre de la organización.'
      if (!orgDocumento) e.orgDocumento = 'Cargue el documento de respaldo.'
    }
    setErrores(e)
    if (Object.keys(e).length) return e
    await guardar({ cargo, respaldo_org: respaldo, org_nombre: respaldo ? orgNombre.trim() : null, org_documento: respaldo ? orgDocumento : null })
  }

  return (
    <Seccion titulo="Selección del cargo al que postula" onGuardar={onGuardar} soloLectura={soloLectura} textoBoton="Guardar postulación">
      <fieldset aria-describedby={errores.cargo ? 'cargo-error' : undefined}>
        <legend className="mb-2 text-sm font-medium">
          Cargo <span className="text-red-600">*</span>
        </legend>
        <div className="flex flex-col gap-2">
          {CARGOS.map((c) => (
            <label key={c.codigo} className="flex items-start gap-2 rounded-md border border-slate-200 p-3 text-sm has-checked:border-emerald-600 has-checked:bg-emerald-50">
              <input type="radio" name="cargo" value={c.codigo} checked={cargo === c.codigo} onChange={() => setCargo(c.codigo)} className="mt-1" />
              {c.nombre}
            </label>
          ))}
        </div>
        {errores.cargo && (
          <p id="cargo-error" role="alert" className="mt-1 text-sm text-red-700">
            {errores.cargo}
          </p>
        )}
      </fieldset>

      {validando && <p className="text-sm text-slate-500">Validando con el Historial Laboral del IESS…</p>}
      <Alerta>{validacion.cargo === cargo && validacion.error}</Alerta>
      {habilitado && (
        <>
          <Alerta tipo="info">{datos.habilitado}</Alerta>
          <SiNo id="respaldo" pregunta={datos.respaldo} value={respaldo} onChange={setRespaldo} error={errores.respaldo} />
          {respaldo && (
            <div className="grid gap-4">
              <Campo id="org-nombre" etiqueta="Nombre de Organización:" maxLength={200} required value={orgNombre} onChange={(e) => setOrgNombre(e.target.value)} error={errores.orgNombre} />
              <CampoPDF id="org-documento" etiqueta="Documento de respaldo" tipo="ORGANIZACION" limite={LIMITE_ORGANIZACION} obligatorio value={orgDocumento} onChange={setOrgDocumento} error={errores.orgDocumento} />
            </div>
          )}
        </>
      )}
    </Seccion>
  )
}
