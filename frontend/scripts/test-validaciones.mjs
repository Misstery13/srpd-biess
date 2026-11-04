// Comprueba los formatos de src/utils/validaciones.js (mismos casos que backend/internal/service/campos_test.go)
// Uso: node scripts/test-validaciones.mjs
import assert from 'node:assert/strict'
import * as v from '../src/utils/validaciones.js'

const casos = {
  esNombre: [['María José Pérez', "Juan O'Neil", 'Ana-Lucía Núñez'], ['Juan123', '12345', 'A', 'Juan@', '<script>']],
  esTexto: [['Universidad Central del Ecuador', 'Av. 10 de Agosto N34-120', 'Banco 24 de Mayo S.A.'], ['12345', '!!!', '<b>Hola</b>', 'a=b', '100%']],
  esNumeroCasa: [['N34-120', 'S/N', 'sn', '12'], ['casa', '#$%']],
  esSenescyt: [['1005-2018-1234567', 'ABC123'], ['abc', '12 34', '12/34']],
}
for (const [fn, [ok, mal]] of Object.entries(casos)) {
  for (const t of ok) assert.ok(v[fn](t), `${fn}(${t}) debía ser válido`)
  for (const t of mal) assert.ok(!v[fn](t), `${fn}(${t}) debía ser inválido`)
}
assert.ok(v.esCelular('0991234567', true) && !v.esCelular('991234567', true) && !v.esCelular('0291234567', true) && v.esCelular('+573001234567', false))
assert.ok(v.esTelefonoFijo('022345678', true) && !v.esTelefonoFijo('0991234567', true) && !v.esTelefonoFijo('abc', false))
assert.equal(v.sinDigitos('Ana 123 Pérez'), 'Ana  Pérez')
assert.equal(v.soloTelefono('+59 3a9-9+1'), '+593991')
assert.ok(v.fechaEnRango('2020-01-01', '2026-11-03') && !v.fechaEnRango('1899-12-31', '2026-11-03') && !v.fechaEnRango('2027-01-01', '2026-11-03'))
const e = {}
v.revisarTexto(e, 'a', '', { vacio: 'obligatorio' })
v.revisarTexto(e, 'b', '12345')
v.revisarTexto(e, 'c', 'Ana 1', { formato: 'nombre' })
v.revisarTexto(e, 'd', '')
assert.deepEqual(Object.keys(e), ['a', 'b', 'c'])
console.log('validaciones: OK')
