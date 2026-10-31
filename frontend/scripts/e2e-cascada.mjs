// Prueba E2E de la cascada País → Provincia → Ciudad con Chrome headless vía CDP.
// Uso: node e2e-cascada.mjs <ruta-chrome> <url>
import { spawn } from 'node:child_process'
import { mkdtempSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const [chrome, url] = process.argv.slice(2)
const perfil = mkdtempSync(join(tmpdir(), 'cdp-'))
const proc = spawn(chrome, ['--headless=new', '--remote-debugging-port=9333', `--user-data-dir=${perfil}`, 'about:blank'])

const esperar = (ms) => new Promise((r) => setTimeout(r, ms))
let ws
for (let i = 0; i < 50 && !ws; i++) {
  try {
    const tabs = await (await fetch('http://127.0.0.1:9333/json')).json()
    ws = tabs.find((t) => t.type === 'page')?.webSocketDebuggerUrl
  } catch {}
  if (!ws) await esperar(200)
}

const sock = new WebSocket(ws)
await new Promise((r) => (sock.onopen = r))
let id = 0
const pend = new Map()
const errores = []
sock.onmessage = (m) => {
  const d = JSON.parse(m.data)
  if (d.id && pend.has(d.id)) pend.get(d.id)(d.result)
  if (d.method === 'Runtime.exceptionThrown') errores.push(d.params.exceptionDetails.text)
  if (d.method === 'Runtime.consoleAPICalled' && d.params.type === 'error') errores.push(d.params.args.map((a) => a.value).join(' '))
}
const cdp = (method, params = {}) =>
  new Promise((r) => {
    pend.set(++id, r)
    sock.send(JSON.stringify({ id, method, params }))
  })
const js = async (expr) => (await cdp('Runtime.evaluate', { expression: expr, awaitPromise: true, returnByValue: true })).result.value

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

await cdp('Runtime.enable')
await cdp('Page.enable')
await cdp('Page.navigate', { url })
await esperar(2500)

let fallos = 0
const ok = (cond, msg) => {
  console.log(`${cond ? 'OK   ' : 'FALLA'} ${msg}`)
  if (!cond) fallos++
}

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
  await cdp('Page.navigate', { url: new URL(r, url).href })
  await esperar(800)
  const h1 = await js(`document.querySelector('h1')?.textContent`)
  ok(!!h1, `ruta ${r} → "${h1}"`)
}

ok(errores.length === 0, `sin errores en consola ${errores.length ? JSON.stringify(errores) : ''}`)
sock.close()
proc.kill()
process.exit(fallos ? 1 : 0)
