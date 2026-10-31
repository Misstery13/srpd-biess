// Prueba E2E de EF-11 (registro, bloqueo, login, recuperación, rutas protegidas) con VITE_MOCK=true.
// Uso: node e2e-auth.mjs <ruta-chrome> <url base, p. ej. http://localhost:5173>
import { abrirChrome, esperar } from './cdp.mjs'

const [chrome, base] = process.argv.slice(2)
const { js, ir, ok, terminar } = await abrirChrome(chrome)
const url = (r) => new URL(r, base).href

// Escribe como un usuario (setter nativo + evento input, para que React lo registre)
const escribir = (sel, valor) => js(`(() => {
  const el = document.querySelector('${sel}')
  Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(el, ${JSON.stringify(valor)})
  el.dispatchEvent(new Event('input', { bubbles: true })) })()`)
const clic = async (sel, ms = 600) => {
  await js(`document.querySelector('${sel}').click()`)
  await esperar(ms)
}
const texto = (sel) => js(`document.querySelector('${sel}')?.textContent ?? null`)
const ruta = () => js('location.pathname')
const alerta = () => texto('form [role=alert]')
const enviar = 'button[type=submit]'

// Rutas protegidas
await ir(url('/postulacion'))
ok((await ruta()) === '/login', 'sin sesión, /postulacion redirige a /login')

// Registro: validación local (V1, V2) sin llamar al backend
await ir(url('/registro'))
await escribir('#cedula', '1710034066')
await escribir('#fecha', '31/02/2020')
await clic(enviar, 200)
ok((await texto('#cedula-error')) !== null, 'cédula con dígito verificador inválido muestra error en el campo')
ok((await texto('#fecha-error')) !== null, 'fecha inexistente (31/02) muestra error en el campo')
await escribir('#fecha', '15032020')
ok((await js(`document.querySelector('#fecha').value`)) === '15/03/2020', 'la fecha se formatea como DD/MM/AAAA')

// Registro: 3 intentos fallidos → bloqueo (S4)
await escribir('#cedula', '1710034065')
await escribir('#fecha', '01/01/2020')
await clic(enviar)
ok((await alerta()) === 'la información ingresada es incorrecta', 'intento 1: mensaje exacto del documento')
await clic(enviar)
ok((await alerta()) === 'la información ingresada es incorrecta', 'intento 2: mensaje exacto del documento')
await clic(enviar)
ok((await alerta()) === 'registro fallido, inténtelo nuevamente en 10 minutos', 'intento 3: mensaje exacto de bloqueo')
ok(await js(`document.querySelector('${enviar}').disabled`), 'botón deshabilitado durante el bloqueo')

// Registro correcto con código dactilar (otra cédula; la anterior quedó bloqueada)
await ir(url('/registro'))
await escribir('#cedula', '0912345675')
await js(`document.querySelectorAll('input[name=metodo]')[1].click()`)
await esperar(100)
await escribir('#codigo', 'e1234i5678')
ok((await js(`document.querySelector('#codigo').value`)) === 'E1234I5678', 'código dactilar en mayúsculas')
await clic(enviar)
ok((await texto('[role=status]'))?.includes('ANA LUCÍA'), 'paso 2 muestra el nombre del Registro Civil')
await escribir('#email', 'ana@example.com')
await escribir('#password', 'clave123')
await escribir('#confirmar', 'clave124')
await clic(enviar, 200)
ok((await texto('#confirmar-error')) !== null, 'contraseñas distintas muestran error')
await escribir('#password', 'corta')
await escribir('#confirmar', 'corta')
await clic(enviar, 200)
ok((await texto('#password-error')) !== null, 'contraseña débil rechazada (S3)')
await escribir('#password', 'clave123')
await escribir('#confirmar', 'clave123')
await clic(enviar)
ok((await ruta()) === '/login' && (await texto('[role=status]'))?.startsWith('Registro exitoso'), 'registro correcto → /login con aviso')

// Login
await escribir('#cedula', '0912345675')
await escribir('#password', 'incorrecta1')
await clic(enviar)
ok((await alerta()) === 'sus credenciales de ingreso son incorrectas', 'login fallido: mensaje exacto del documento')
ok((await js(`[...document.querySelectorAll('a')].some(a => a.textContent === '¿Olvidaste tu contraseña?')`)), 'link "¿Olvidaste tu contraseña?"')
await escribir('#password', 'clave123')
await clic(enviar, 1500)
ok((await ruta()) === '/postulacion', 'login correcto → /postulacion (primer ingreso, EF-12)')
ok((await js(`!!sessionStorage.getItem('token')`)), 'token guardado en sessionStorage')
await ir(url('/login'))
ok((await ruta()) === '/postulacion', 'con sesión, /login redirige a /postulacion')
await clic('header button')
ok((await ruta()) === '/login' && !(await js(`sessionStorage.getItem('token')`)), 'cerrar sesión borra el token y vuelve a /login')

// E3: un 401 con sesión iniciada (token vencido) la cierra y vuelve a /login. Se simula la respuesta con un adaptador.
// (Page.navigate recarga la página y el mock pierde las cuentas, así que se usa un token directo)
await js(`sessionStorage.setItem('token', 'vencido')`)
await ir(url('/postulacion'))
ok((await ruta()) === '/postulacion', 'con token, /postulacion es accesible')
await js(`(async () => {
  const { default: api } = await import('/src/services/api.js')
  api.defaults.adapter = (config) => Promise.reject({ config, response: { status: 401, data: {} } })
  await api.get('/postulacion').catch(() => {})
})()`)
await esperar(500)
ok((await ruta()) === '/login' && (await texto('[role=status]'))?.includes('sesión terminó'), '401 con sesión → cierra sesión, /login con aviso')
ok(!(await js(`sessionStorage.getItem('token')`)), '401 con sesión → token borrado')

// Recuperación y restablecimiento
await ir(url('/recuperar'))
await escribir('#cedula', '0912345675')
await escribir('#fecha', '02/11/2019')
await clic(enviar)
ok((await texto('[role=status]'))?.includes('correo'), 'recuperar: confirma el envío del correo')
await ir(url('/restablecer'))
ok((await alerta()) !== null && (await js(`document.querySelector('${enviar}').disabled`)), 'restablecer sin token: aviso y formulario deshabilitado')
await ir(url('/restablecer?token=caducado'))
await escribir('#password', 'nueva1234')
await escribir('#confirmar', 'nueva1234')
await clic(enviar)
ok((await alerta()) !== null, 'token inválido: muestra el error del servidor')
await ir(url('/restablecer?token=mock'))
await escribir('#password', 'nueva1234')
await escribir('#confirmar', 'nueva1234')
await clic(enviar)
ok((await ruta()) === '/login' && (await texto('[role=status]'))?.includes('restablecida'), 'restablecer correcto → /login con aviso')

terminar()
