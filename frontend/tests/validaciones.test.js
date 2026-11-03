import { describe, it } from 'node:test'
import assert from 'node:assert/strict'
import {
  cedulaValida,
  passwordValida,
  codigoDactilarValido,
  emailValido,
  formatearFecha,
  fechaAIso,
  validarIdentidad,
} from '../src/utils/validaciones.js'

describe('Validaciones Frontend — Reglas de Autenticación y Registro (EF-11)', () => {
  describe('cedulaValida (Algoritmo Módulo 10)', () => {
    it('debe aceptar las cédulas de prueba válidas del README', () => {
      assert.equal(cedulaValida('1710034065'), true, 'Juan Pérez (Pichincha) debe ser válida')
      assert.equal(cedulaValida('0912345675'), true, 'Ana Mora (Guayas) debe ser válida')
      assert.equal(cedulaValida('0102030400'), true, 'Pedro Salazar (Azuay) debe ser válida')
      assert.equal(cedulaValida('1754321980'), true, 'Carlos Gómez (Pichincha) debe ser válida')
    })

    it('debe rechazar cédulas con longitud diferente de 10 dígitos', () => {
      assert.equal(cedulaValida('171003406'), false, 'Menos de 10 dígitos debe fallar')
      assert.equal(cedulaValida('17100340655'), false, 'Más de 10 dígitos debe fallar')
      assert.equal(cedulaValida(''), false, 'Cédula vacía debe fallar')
    })

    it('debe rechazar cédulas con código de provincia inválido', () => {
      assert.equal(cedulaValida('0010034065'), false, 'Provincia 00 no existe')
      assert.equal(cedulaValida('2510034065'), false, 'Provincia 25 no existe en Ecuador')
      assert.equal(cedulaValida('9910034065'), false, 'Provincia 99 no existe')
    })

    it('debe rechazar cédulas de personas naturales con tercer dígito >= 6', () => {
      assert.equal(cedulaValida('1760034065'), false, 'Tercer dígito 6 no es persona natural')
      assert.equal(cedulaValida('1790034065'), false, 'Tercer dígito 9 es persona jurídica')
    })

    it('debe rechazar cédulas donde falle el dígito verificador módulo 10', () => {
      // 1710034065 es válida, si cambiamos el último dígito a 4 debe fallar
      assert.equal(cedulaValida('1710034064'), false, 'Dígito verificador alterado debe fallar')
      assert.equal(cedulaValida('0912345671'), false, 'Dígito verificador alterado debe fallar')
    })

    it('debe rechazar cédulas con caracteres no numéricos', () => {
      assert.equal(cedulaValida('171003406A'), false)
      assert.equal(cedulaValida('17100 4065'), false)
    })
  })

  describe('passwordValida (Decisión S3)', () => {
    it('debe exigir al menos 8 caracteres, al menos 1 letra y al menos 1 número', () => {
      assert.equal(passwordValida('Clave123'), true, '8 caracteres con letras y números es válida')
      assert.equal(passwordValida('superpassword2026!'), true, 'Más de 8 caracteres es válida')
    })

    it('debe rechazar contraseñas que no cumplan los requisitos mínimos', () => {
      assert.equal(passwordValida('Abc1'), false, 'Menos de 8 caracteres')
      assert.equal(passwordValida('sololetras'), false, 'Sin números')
      assert.equal(passwordValida('12345678'), false, 'Sin letras')
      assert.equal(passwordValida(''), false, 'Vacía')
    })
  })

  describe('codigoDactilarValido (Decisión V3)', () => {
    it('debe aceptar códigos alfanuméricos de 10 caracteres en mayúsculas', () => {
      assert.equal(codigoDactilarValido('V4443V4444'), true)
      assert.equal(codigoDactilarValido('E1234I5678'), true)
      assert.equal(codigoDactilarValido('A1111B2222'), true)
    })

    it('debe rechazar códigos con minúsculas, caracteres especiales o longitud incorrecta', () => {
      assert.equal(codigoDactilarValido('v4443v4444'), false, 'Debe ser en mayúsculas')
      assert.equal(codigoDactilarValido('V4443V444'), false, '9 caracteres')
      assert.equal(codigoDactilarValido('V4443V44445'), false, '11 caracteres')
      assert.equal(codigoDactilarValido('V4443-4444'), false, 'Con guion')
    })
  })

  describe('emailValido', () => {
    it('debe aceptar correos electrónicos bien formados', () => {
      assert.equal(emailValido('juan.perez@biess.fin.ec'), true)
      assert.equal(emailValido('postulante@gmail.com'), true)
    })

    it('debe rechazar formatos de correo incorrectos o con espacios', () => {
      assert.equal(emailValido('correo-sin-arroba.com'), false)
      assert.equal(emailValido('correo@'), false)
      assert.equal(emailValido('correo@dominio'), false)
      assert.equal(emailValido('correo @dominio.com'), false)
    })
  })

  describe('formatearFecha y fechaAIso (Decisión V2)', () => {
    it('formatearFecha debe agregar las barras DD/MM/AAAA conforme se escribe', () => {
      assert.equal(formatearFecha('15032020'), '15/03/2020')
      assert.equal(formatearFecha('1503'), '15/03')
      assert.equal(formatearFecha('15'), '15')
    })

    it('fechaAIso debe convertir DD/MM/AAAA al formato estándar YYYY-MM-DD', () => {
      assert.equal(fechaAIso('15/03/2020'), '2020-03-15')
      assert.equal(fechaAIso('02/11/2019'), '2019-11-02')
    })

    it('fechaAIso debe rechazar fechas imposibles o futuras', () => {
      assert.equal(fechaAIso('31/02/2020'), null, 'Febrero 31 no existe')
      assert.equal(fechaAIso('32/01/2020'), null, 'Día 32 no existe')
      assert.equal(fechaAIso('15/13/2020'), null, 'Mes 13 no existe')
      assert.equal(fechaAIso('01/01/2099'), null, 'Fecha futura no es válida para expedición')
    })
  })

  describe('validarIdentidad', () => {
    it('debe retornar datos limpios y sin errores cuando la identidad es válida', () => {
      const res = validarIdentidad({
        cedula: '1710034065',
        metodo: 'fecha',
        fecha: '15/03/2020',
        codigo: '',
      })
      assert.deepEqual(res.errores, {})
      assert.equal(res.datos.cedula, '1710034065')
      assert.equal(res.datos.fecha_expedicion, '2020-03-15')
    })

    it('debe retornar mensaje de error si la cédula es inválida', () => {
      const res = validarIdentidad({
        cedula: '1710034064', // Último dígito inválido
        metodo: 'codigo',
        fecha: '',
        codigo: 'V4443V4444',
      })
      assert.equal(typeof res.errores.cedula, 'string')
    })
  })
})

