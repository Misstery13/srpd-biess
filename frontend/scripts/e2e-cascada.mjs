// Prueba E2E de la cascada País → Provincia → Ciudad con Chrome headless vía CDP.
// Uso: node e2e-cascada.mjs <ruta-chrome> <url de /postulacion>
import { abrirChrome, esperar } from './cdp.mjs'

const [chrome, url] = process.argv.slice(2)
const { js, ir, ok, terminar } = await abrirChrome(chrome)

// Elige una opción por texto como lo haría React (setter nativo + evento change)
const elegir = (sel, texto) => js(`(() => {
  const s = document.querySelector('${sel}')
  const o = [...s.options].find(o => o.text === '${texto}')
  Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, o.value)
  s.dispatchEvent(new Event('change', { bubbles: true }))
  return o.value })()`)
const opciones = (sel) => js(`[...document.querySelector('${sel}')?.options ?? []].map(o => o.text)`)
const existe = (sel) => js(`!!document.querySelector('${sel}')`)
const deshab = (sel) => js(`document.querySelector('${sel}').disabled`)

// /postulacion es ruta protegida: se simula una sesión iniciada (los catálogos son públicos)
await ir(new URL('/', url).href, 500)
await js(`sessionStorage.setItem('token', 'e2e')`)
await ir(url, 2500)

ok((await opciones('#pais')).includes('Ecuador'), 'países cargados')
ok(await deshab('#provincia'), 'provincia deshabilitada sin país')
await elegir('#pais', 'Ecuador')
ok((await opciones('#provincia'))[0] === 'Cargando…', 'loader mientras carga provincias')
await esperar(600)
ok((await opciones('#provincia')).includes('Pichincha'), 'provincias de Ecuador')
ok(await deshab('#canton'), 'ciudad deshabilitada sin provincia')
await elegir('#provincia', 'Pichincha')
await esperar(600)
ok((await opciones('#canton')).includes('Quito'), 'cantones de Pichincha')
await elegir('#canton', 'Quito')
await elegir('#provincia', 'Guayas')
await esperar(600)
const c = await opciones('#canton')
ok(c.includes('Guayaquil') && !c.includes('Quito'), 'cambiar provincia recarga cantones')
ok((await js(`document.querySelector('#canton').value`)) === '', 'cambiar provincia limpia la ciudad')
await elegir('#pais', 'Colombia')
await esperar(400)
ok((await existe('#provinciaTexto')) && (await existe('#ciudadTexto')) && !(await existe('#provincia')), 'país extranjero: provincia y ciudad en texto libre')
await elegir('#pais', 'Ecuador')
await esperar(600)
ok((await js(`document.querySelector('#provincia').value`)) === '', 'volver a Ecuador limpia la provincia')

// Rutas
for (const r of ['/', '/login', '/registro', '/recuperar', '/no-existe']) {
  await ir(new URL(r, url).href, 800)
  const h1 = await js(`document.querySelector('h1')?.textContent`)
  ok(!!h1, `ruta ${r} → "${h1}"`)
}

terminar()
