// Chrome headless controlado por CDP, sin dependencias. Lo usan los scripts e2e-*.mjs.
import { spawn } from 'node:child_process'
import { mkdtempSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

export const esperar = (ms) => new Promise((r) => setTimeout(r, ms))

export async function abrirChrome(chrome) {
  const perfil = mkdtempSync(join(tmpdir(), 'cdp-'))
  const proc = spawn(chrome, ['--headless=new', '--remote-debugging-port=9333', `--user-data-dir=${perfil}`, 'about:blank'])

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

  await cdp('Runtime.enable')
  await cdp('Page.enable')

  let fallos = 0
  const ok = (cond, msg) => {
    console.log(`${cond ? 'OK   ' : 'FALLA'} ${msg}`)
    if (!cond) fallos++
  }
  return {
    cdp,
    js,
    ok,
    ir: async (url, ms = 1200) => {
      await cdp('Page.navigate', { url })
      await esperar(ms)
    },
    terminar() {
      ok(errores.length === 0, `sin errores en consola ${errores.length ? JSON.stringify(errores) : ''}`)
      sock.close()
      proc.kill()
      process.exit(fallos ? 1 : 0)
    },
  }
}
