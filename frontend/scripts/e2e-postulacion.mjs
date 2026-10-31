// Prueba E2E de EF-12 (formulario de postulación) con VITE_MOCK=true.
// Uso: node e2e-postulacion.mjs <ruta-chrome> <url base, p. ej. http://localhost:5173>
import { writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { abrirChrome, esperar } from './cdp.mjs'

const [chrome, base] = process.argv.slice(2)
const { cdp, js, ir, ok, terminar } = await abrirChrome(chrome)
const url = (r) => new URL(r, base).href

// PDFs de prueba
const pdf = join(tmpdir(), 'e2e-respaldo.pdf')
const noPdf = join(tmpdir(), 'e2e-foto.png')
const grande = join(tmpdir(), 'e2e-grande.pdf')
writeFileSync(pdf, '%PDF-1.4\n%%EOF')
writeFileSync(noPdf, 'PNG')
writeFileSync(grande, Buffer.concat([Buffer.from('%PDF-1.4\n'), Buffer.alloc(2.5 * 1024 * 1024)]))

// --- utilidades: actúan como un usuario (setters nativos + eventos, para que React los registre)
const q = (sel) => JSON.stringify(sel)
const escribir = (sel, valor) => js(`(() => {
  const el = document.querySelector(${q(sel)})
  const proto = el instanceof HTMLTextAreaElement ? HTMLTextAreaElement.prototype : HTMLInputElement.prototype
  Object.getOwnPropertyDescriptor(proto, 'value').set.call(el, ${q(valor)})
  el.dispatchEvent(new Event('input', { bubbles: true })) })()`)
const elegir = (sel, texto) => js(`(() => {
  const s = document.querySelector(${q(sel)})
  const o = [...s.options].find(o => o.text === ${q(texto)})
  Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, o.value)
  s.dispatchEvent(new Event('change', { bubbles: true })) })()`)
const clic = async (sel, ms = 400) => {
  await js(`document.querySelector(${q(sel)}).click()`)
  await esperar(ms)
}
// Botón por su texto dentro de un contenedor
const boton = async (scope, texto, ms = 500) => {
  const hecho = await js(`(() => { const b = [...document.querySelectorAll(${q(scope + ' button')})].find(b => b.textContent.trim() === ${q(texto)}); if (!b) return false; b.click(); return true })()`)
  if (!hecho) console.log(`   (no se encontró el botón "${texto}" en ${scope})`)
  await esperar(ms)
}
// Radio Si/No de un grupo por name
const siNo = async (name, si) => {
  await js(`document.querySelectorAll('input[name="${name}"]')[${si ? 0 : 1}].click()`)
  await esperar(150)
}
const texto = (sel) => js(`document.querySelector(${q(sel)})?.textContent ?? null`)
const existe = (sel) => js(`!!document.querySelector(${q(sel)})`)
const cuenta = (sel) => js(`document.querySelectorAll(${q(sel)}).length`)
const pestana = (titulo) => boton('nav', titulo, 300)
async function subir(sel, archivo) {
  const { root } = await cdp('DOM.getDocument')
  const { nodeId } = await cdp('DOM.querySelector', { nodeId: root.nodeId, selector: sel })
  await cdp('DOM.setFileInputFiles', { nodeId, files: [archivo] })
  await esperar(600)
}
async function sesion(cedula, extra = '') {
  await ir(url('/'), 300)
  await js(`sessionStorage.clear(); sessionStorage.setItem('token', 'mock.${cedula}'); ${extra}`)
  await ir(url('/postulacion'), 1500)
}

// ===== A. Postulante afiliado y casado (1710034065): recorrido completo =====
await sesion('1710034065')
ok((await texto('[data-personal]'))?.includes('JUAN CARLOS'), '1.1 datos del Registro Civil cargados')
ok((await texto('[data-personal]'))?.includes('MARÍA ELENA TORRES'), '1.1.7.1 casado: se muestran los datos de la pareja')
ok(!(await existe('[data-personal] input')), '1.1 información personal en solo lectura (P6)')

// Domicilio
const D = '[data-seccion=domicilio]'
await pestana('Domicilio y contacto')
await boton(D, 'Guardar progreso')
ok((await cuenta(`${D} [role=alert]`)) >= 3, '1.2 guardar vacío: errores junto a los campos (E5)')
await elegir('#pais', 'Ecuador')
await esperar(500)
await elegir('#provincia', 'Pichincha')
await esperar(500)
await elegir('#canton', 'Quito')
await escribir('#celular', '0991234567')
await escribir('#calle_principal', 'Av. Amazonas')
ok((await js(`document.querySelector('#email1').value`)) === '1710034065@example.com', '1.2.5 correo 1 = correo registrado (solo lectura)')
await boton(D, 'Guardar progreso', 600)
ok((await texto(`${D} output`)) === 'Información guardada.', '1.2.13 domicilio guardado')
ok((await texto('[data-codigo]')) === 'BIESS-2026-000001', 'código único de postulación visible')

// Información adicional (familiares)
const F = '[data-seccion=adicional]'
await pestana('Información adicional')
await boton(F, 'Guardar información')
ok(!!(await texto('#tiene-familiares-error')), '1.3.1 exige responder Si/No')
await siNo('tiene-familiares', true)
await boton(F, '+ Agregar')
await boton(F, 'Agregar a la lista')
ok((await cuenta(`${F} [data-edicion] [role=alert]`)) === 3, '1.3.1 familiar vacío: 3 campos obligatorios')
ok((await js(`[...document.querySelector('#fam-parentesco').options].length`)) === 16, '1.3.1.2 lista de 15 parentescos')
await escribir('#fam-nombres', 'Pedro Pérez')
await elegir('#fam-parentesco', 'hermano/a')
await elegir('#fam-institucion', 'BIESS')
await escribir('#fam-cargo', 'Analista')
await boton(F, 'Agregar a la lista')
ok((await cuenta(`${F} li`)) === 1, '1.3.1.6 Agregar')
await boton(F, 'Editar')
await escribir('#fam-cargo', 'Director')
await boton(F, 'Aplicar cambios')
ok((await texto(`${F} li`))?.includes('Director'), '1.3.1.7 Editar')
await boton(F, 'Eliminar')
ok((await cuenta(`${F} li`)) === 0, '1.3.1.7 Eliminar')
await siNo('tiene-familiares', false)
await boton(F, 'Guardar información', 600)
ok((await texto(`${F} output`)) === 'Información guardada.', '1.3.1.7 Guardar información (No)')

// Cargo
const C = '[data-seccion=cargo]'
await pestana('Cargo')
await clic('input[name=cargo][value=AFILIADOS]', 600)
ok((await texto(`${C} output`)) === 'Postulante afiliado al IESS', '1.4.1.1 afiliado: "Postulante afiliado al IESS"')
ok((await texto(`${C} fieldset legend`)) !== null && (await js(`document.querySelector('${C}').textContent.includes('organización de trabajadores')`)), '1.4.1.3 pregunta de respaldo de organización de trabajadores')
await siNo('respaldo', true)
await boton(C, 'Guardar postulación')
ok(!!(await texto('#org-nombre-error')) && !!(await texto('#org-documento-error')), '1.4.1.3.1 nombre y documento de organización obligatorios (*)')
await subir('#org-documento', noPdf)
ok((await texto('#org-documento-error'))?.includes('PDF'), 'A2 archivo que no es PDF rechazado en el navegador')
await escribir('#org-nombre', 'Federación Nacional de Trabajadores')
await subir('#org-documento', pdf)
ok(await existe(`${C} [data-documento]`), '1.4.1.3.1.2 documento de respaldo cargado')
await boton(C, 'Guardar postulación', 600)
ok((await texto(`${C} output:last-of-type`)) === 'Información guardada.' || (await js(`[...document.querySelectorAll('${C} output')].some(o => o.textContent === 'Información guardada.')`)), '1.4.1.4 cargo guardado')

// Formación
const FO = '[data-seccion=formacion]'
await pestana('Formación académica')
await boton(FO, '+ Agregar')
ok((await js(`[...document.querySelector('#for-nivel').options].map(o => o.text).slice(1).join('|')`)) === 'Tercer Nivel de Grado|Cuarto Nivel - Especialidad Académica|Cuarto Nivel - Maestría Académica|Cuarto Nivel PhD - Doctorado', '1.5.1 niveles del documento')
await elegir('#for-nivel', 'Cuarto Nivel - Maestría Académica')
await escribir('#for-institucion', 'Universidad Central')
await elegir('#for-tipo', 'Nacional')
await escribir('#for-titulo', 'Magíster en Finanzas')
await elegir('#for-area', 'Finanzas')
await subir('#for-documento', grande)
ok((await texto('#for-documento-error'))?.includes('2 MB'), 'A3 PDF de formación > 2 MB rechazado')
await subir('#for-documento', pdf)
await boton(FO, 'Agregar a la lista')
await boton(FO, 'Guardar información', 600)
ok((await cuenta(`${FO} li`)) === 1 && (await texto(`${FO} output`)) === 'Información guardada.', '1.5 formación agregada y guardada')

// Experiencia y calculadora
const E = '[data-seccion=experiencia]'
await pestana('Experiencia')
await boton(E, '+ Agregar')
await elegir('#exp-tipo', 'Dirección')
await escribir('#exp-cargo', 'Gerente financiero')
await escribir('#exp-institucion', 'Empresa XYZ')
await elegir('#exp-area', 'Finanzas')
await escribir('#exp-inicio', '2020-01-15')
await escribir('#exp-fin', '2019-01-01')
await boton(E, 'Agregar a la lista')
ok((await texto('#exp-fin-error'))?.includes('anterior'), 'V7 fecha fin anterior a inicio rechazada')
ok(!!(await texto('#exp-documento-error')), '1.6.8 documento de respaldo obligatorio (*)')
await escribir('#exp-fin', '2023-03-20')
ok((await texto(`${E} [data-edicion] [data-duracion]`))?.includes('3 años, 2 meses y 5 días'), '1.6 calculadora: 3 años, 2 meses y 5 días')
await subir('#exp-documento', pdf)
ok((await js(`document.querySelector('#exp-funciones').maxLength`)) === 1000, 'V8 funciones con máximo de 1000 caracteres al escribir')
await escribir('#exp-funciones', 'x'.repeat(1001))
await boton(E, 'Agregar a la lista')
ok((await texto('#exp-funciones-error'))?.includes('1000'), 'V8 más de 1000 caracteres rechazado')
await escribir('#exp-funciones', 'Gestión del presupuesto y la tesorería.')
await boton(E, 'Agregar a la lista')
await boton(E, 'Guardar información', 600)
ok((await texto(`${E} [data-total]`))?.includes('3 años, 2 meses y 5 días'), '1.6 total de experiencia')

// Publicaciones y capacitación (respuesta "No" / lista vacía)
await pestana('Publicaciones')
await siNo('tiene-publicaciones', false)
await boton('[data-seccion=publicaciones]', 'Guardar información', 600)
ok((await texto('[data-seccion=publicaciones] output')) === 'Información guardada.', '1.7 publicaciones (No) guardado')
await pestana('Capacitación')
await boton('[data-seccion=capacitacion]', '+ Agregar')
await escribir('#cap-horas', '0')
await boton('[data-seccion=capacitacion]', 'Agregar a la lista')
ok((await cuenta('[data-seccion=capacitacion] [data-edicion] [role=alert]')) === 6, '1.8 capacitación: 6 obligatorios y horas > 0')
await boton('[data-seccion=capacitacion]', 'Cancelar')

// Envío bloqueado sin inhabilidades
const V = '[data-seccion=envio]'
await pestana('Guardar y enviar')
ok((await texto(`${V} [data-pendientes]`))?.includes('Inhabilidades'), 'envío: avisa que faltan las inhabilidades')

// Inhabilidades
const I = '[data-seccion=inhabilidades]'
await pestana('Inhabilidades')
ok((await js(`[...document.querySelectorAll('${I} input[type=radio]')].every(r => !r.checked)`)), '1.9 ninguna respuesta marcada por defecto (declaración jurada)')
await boton(I, 'Guardar información')
ok((await cuenta(`${I} [role=alert]`)) === 14, '1.9 las 13 preguntas son obligatorias')
const claves = await js(`[...new Set([...document.querySelectorAll('${I} input[type=radio]')].map(r => r.name))]`)
for (const c of claves) await js(`document.querySelectorAll('input[name="${c}"]')[${c === 'goce_derechos_politicos' || c === 'removido_control' ? 0 : 1}].click()`)
await esperar(200)
await boton(I, 'Guardar información')
ok(!!(await texto('#removido_institucion-error')) && !!(await texto('#removido_motivo-error')), '1.9.7.1 removido: institución y motivo obligatorios')
await escribir('#removido_institucion', 'Superintendencia de Bancos')
await escribir('#removido_motivo', 'Reestructuración')
await boton(I, 'Guardar información', 600)
ok((await texto(`${I} output`)) === 'Información guardada.', '1.9 inhabilidades guardadas')

// Guardar y enviar con consentimiento
await pestana('Guardar y enviar')
ok(!(await existe(`${V} [data-pendientes]`)), 'envío: sin secciones pendientes')
ok(!(await existe(`${V} [data-consentimiento]`)), '1.10.1 consentimiento aparece solo al pulsar "Guardar Postulación"')
await boton(V, 'Guardar Postulación', 600)
ok((await cuenta(`${V} [data-consentimiento] input[type=checkbox]`)) === 5, '1.10.1 los 5 textos de consentimiento')
ok((await js(`[...document.querySelectorAll('${V} [data-consentimiento] input')].every(c => !c.checked)`)), 'L1 casillas sin marcar por defecto')
ok(await js(`document.querySelector('${V} [data-enviar]').disabled`), 'L1 "Enviar Postulación" deshabilitado sin aceptar')
for (let i = 0; i < 4; i++) await js(`document.querySelectorAll('${V} [data-consentimiento] input')[${i}].click()`)
ok(await js(`document.querySelector('${V} [data-enviar]').disabled`), 'L1 sigue deshabilitado con 4 de 5')
await js(`document.querySelectorAll('${V} [data-consentimiento] input')[4].click()`)
await esperar(100)
await clic(`${V} [data-enviar]`, 700)
ok((await texto(`${V} output`))?.includes('BIESS-2026-000001 fue enviada'), '1.10.3 postulación enviada')
ok(await js(`document.querySelector('[data-seccion=domicilio] fieldset').disabled`), 'P3 enviada: formulario en solo lectura')
ok(await existe('article.print\\:block'), '1.10.2 formulario imprimible con firma, nombres y cédula')
ok((await texto('article'))?.includes('Firma del postulante') && (await texto('article'))?.includes('C.I. 1710034065'), '1.10.2 firma, nombres y cédula en el impreso')

// Eliminar cuenta (L5)
await boton('main', 'Eliminar mi cuenta')
await escribir('#confirmar-eliminar', 'eliminar')
ok(await js(`[...document.querySelectorAll('main button')].find(b => b.textContent === 'Eliminar definitivamente').disabled`), 'L5 exige escribir ELIMINAR')
await escribir('#confirmar-eliminar', 'ELIMINAR')
await boton('main', 'Eliminar definitivamente', 800)
ok((await js('location.pathname')) === '/login' && (await texto('output'))?.includes('eliminados') && !(await js(`sessionStorage.getItem('token')`)), 'L5 cuenta eliminada → /login con aviso y sin sesión')

// ===== B. No ecuatoriano (1754321980) =====
await sesion('1754321980')
ok((await existe('[data-cierre]')) && (await texto('[data-cierre]'))?.includes('nacionalidad ecuatoriana'), '1.1.6 no ecuatoriano: la postulación se cierra')

// ===== C. Jubilada no afiliada (0912345675) elige Afiliados =====
await sesion('0912345675')
await pestana('Cargo')
await clic('input[name=cargo][value=AFILIADOS]', 700)
ok(
  (await texto('[data-cierre]'))?.includes('Postulante No Afiliado al IESS') &&
    (await texto('[data-cierre]'))?.includes('NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS'),
  '1.4.1.2 no afiliado: mensaje exacto y cierre',
)

// ===== D. Período cerrado (T2) =====
await sesion('0912345675', `sessionStorage.setItem('mock:periodo', 'cerrado')`)
ok((await texto('[role=alert]'))?.includes('El período de postulación no está habilitado'), 'T2 aviso de período cerrado')
ok(await js(`document.querySelector('[data-seccion=domicilio] fieldset').disabled`), 'T2 formulario en solo lectura')

terminar()
