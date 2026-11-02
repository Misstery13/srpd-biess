/**
 * @file ef12.test.js
 * @description Pruebas unitarias para utilidades y cálculos de EF-12 (duración, suma de experiencia, cargos)
 */

import { describe, it } from 'node:test';
import assert from 'node:assert/strict';

import {
  calcularDuracion,
  sumarDuraciones,
  cargoPorCodigo,
  tienePareja,
  aFechaApi,
  deFechaApi,
  textoDuracion,
} from '../src/utils/ef12.js';

describe('EF-12: calcularDuracion', () => {
  it('retorna null si falta inicio', () => {
    assert.equal(calcularDuracion(null, '2023-01-01'), null);
    assert.equal(calcularDuracion('', '2023-01-01'), null);
  });

  it('calcula hasta hoy si no se especifica fin (trabajo vigente)', () => {
    const hoyFijo = new Date(2024, 0, 15, 12, 0, 0); // 15 de enero de 2024 local
    const dur = calcularDuracion('2023-01-15', null, hoyFijo);
    assert.deepEqual(dur, { anios: 1, meses: 0, dias: 0 });
  });

  it('retorna null si inicio > fin', () => {
    assert.equal(calcularDuracion('2024-05-01', '2023-05-01'), null);
  });

  it('calcula duración exacta de 1 año', () => {
    const dur = calcularDuracion('2022-01-01', '2023-01-01');
    assert.deepEqual(dur, { anios: 1, meses: 0, dias: 0 });
  });

  it('calcula meses y días correctamente', () => {
    const dur = calcularDuracion('2023-01-10', '2023-03-25');
    assert.deepEqual(dur, { anios: 0, meses: 2, dias: 15 });
  });

  it('ajusta préstamo de días cuando el día de fin es menor al de inicio', () => {
    const dur = calcularDuracion('2023-01-25', '2023-03-10');
    assert.ok(dur !== null);
    assert.equal(dur.anios, 0);
    assert.equal(dur.meses, 1);
    assert.ok(dur.dias > 0);
  });
});

describe('EF-12: sumarDuraciones', () => {
  it('retorna 0 años, meses y días para lista vacía', () => {
    assert.deepEqual(sumarDuraciones([]), { anios: 0, meses: 0, dias: 0 });
  });

  it('suma correctamente y normaliza 30 días a 1 mes', () => {
    const periodos = [
      { anios: 1, meses: 2, dias: 20 },
      { anios: 0, meses: 3, dias: 15 },
    ];
    // 20 + 15 = 35 días => 5 días y 1 mes extra
    // 2 + 3 + 1 = 6 meses
    // 1 año
    assert.deepEqual(sumarDuraciones(periodos), { anios: 1, meses: 6, dias: 5 });
  });

  it('normaliza 12 meses a 1 año', () => {
    const periodos = [
      { anios: 1, meses: 8, dias: 0 },
      { anios: 0, meses: 5, dias: 0 },
    ];
    // 13 meses => 1 mes y 1 año extra => 2 años, 1 mes, 0 días
    assert.deepEqual(sumarDuraciones(periodos), { anios: 2, meses: 1, dias: 0 });
  });

  it('ignora elementos nulos o vacíos', () => {
    const periodos = [
      null,
      { anios: 2, meses: 0, dias: 10 },
      undefined,
    ];
    assert.deepEqual(sumarDuraciones(periodos), { anios: 2, meses: 0, dias: 10 });
  });
});

describe('EF-12: cargoPorCodigo', () => {
  it('resuelve el objeto de configuración oficial para los 3 cargos del concurso', () => {
    const afiliados = cargoPorCodigo('AFILIADOS');
    assert.ok(afiliados);
    assert.equal(afiliados.codigo, 'AFILIADOS');
    assert.match(afiliados.nombre, /Afiliados/i);

    const jubilados = cargoPorCodigo('JUBILADOS');
    assert.ok(jubilados);
    assert.equal(jubilados.codigo, 'JUBILADOS');
    assert.match(jubilados.nombre, /Jubilados/i);

    const empleadores = cargoPorCodigo('EMPLEADORES');
    assert.ok(empleadores);
    assert.equal(empleadores.codigo, 'EMPLEADORES');
    assert.match(empleadores.nombre, /Empleador/i);
  });

  it('retorna undefined si el código no existe', () => {
    assert.equal(cargoPorCodigo('INVALIDO'), undefined);
    assert.equal(cargoPorCodigo(''), undefined);
  });
});

describe('EF-12: tienePareja', () => {
  it('reconoce estados civiles que aplican para cónyuge/pareja', () => {
    assert.equal(tienePareja('CASADO'), true);
    assert.equal(tienePareja('CASADO/A'), true);
    assert.equal(tienePareja('UNION DE HECHO'), true);
    assert.equal(tienePareja('UNIÓN DE HECHO'), true);
  });

  it('retorna false para soltero, divorciado, viudo o vacío', () => {
    assert.equal(tienePareja('SOLTERO'), false);
    assert.equal(tienePareja('DIVORCIADO'), false);
    assert.equal(tienePareja('VIUDO'), false);
    assert.equal(tienePareja(''), false);
    assert.equal(tienePareja(null), false);
  });
});

describe('EF-12: formatos de fecha y texto', () => {
  it('convierte entre fecha ISO y API RFC3339', () => {
    assert.equal(aFechaApi('2023-05-15'), '2023-05-15T00:00:00Z');
    assert.equal(aFechaApi(null), null);
    assert.equal(deFechaApi('2023-05-15T10:30:00Z'), '2023-05-15');
    assert.equal(deFechaApi(null), '');
  });

  it('formatea texto legible de duración', () => {
    assert.equal(textoDuracion({ anios: 1, meses: 1, dias: 1 }), '1 año, 1 mes y 1 día');
    assert.equal(textoDuracion({ anios: 2, meses: 4, dias: 10 }), '2 años, 4 meses y 10 días');
    assert.equal(textoDuracion(null), '—');
  });
});
