import http from 'k6/http';
import { check, group } from 'k6';

// ==============================================================================
// SRPD BIESS — Suite k6 de Broken Access Control & Endpoints Protegidos
// Pista D (Datos & Seguridad) — Responsable: Gino Bermúdez (@ginobermudez76)
// Estándar: OWASP Top 10 (A01: Broken Access Control)
// ==============================================================================

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    checks: ['rate==1.0'], // 100% de aserciones de seguridad deben pasar
  },
};

// Lista de los 7 endpoints estrictamente protegidos según el contrato (Fase 0)
const PROTECTED_ENDPOINTS = [
  { method: 'GET', url: '/api/postulante/me', name: 'GET /api/postulante/me' },
  { method: 'DELETE', url: '/api/postulante/me', name: 'DELETE /api/postulante/me' },
  { method: 'POST', url: '/api/postulacion/validar-cargo', name: 'POST /api/postulacion/validar-cargo', body: JSON.stringify({ cargo: 'AFILIADOS' }) },
  { method: 'GET', url: '/api/postulacion', name: 'GET /api/postulacion' },
  { method: 'PUT', url: '/api/postulacion', name: 'PUT /api/postulacion', body: JSON.stringify({ cargo: 'AFILIADOS' }) },
  { method: 'POST', url: '/api/postulacion/documentos', name: 'POST /api/postulacion/documentos' },
  { method: 'POST', url: '/api/postulacion/enviar', name: 'POST /api/postulacion/enviar', body: JSON.stringify({ acepta_terminos: true }) },
];

// Lista de endpoints que deben ser públicos
const PUBLIC_ENDPOINTS = [
  { method: 'GET', url: '/api/health', name: 'GET /api/health' },
  { method: 'GET', url: '/api/periodo', name: 'GET /api/periodo' },
  { method: 'GET', url: '/api/catalogos/paises', name: 'GET /api/catalogos/paises' },
  { method: 'GET', url: '/api/catalogos/paises/52/provincias', name: 'GET /api/catalogos/paises/:id/provincias' },
  { method: 'GET', url: '/api/catalogos/provincias/17/cantones', name: 'GET /api/catalogos/provincias/:id/cantones' },
];

export default function securityEndpointsTest() {
  console.log(`[k6-Security] Evaluando protección de acceso contra: ${BASE_URL}`);

  // ----------------------------------------------------------------------------
  // TEST 1: PETICIONES ANÓNIMAS (SIN HEADER AUTHORIZATION)
  // Ningún endpoint protegido debe responder 200, 400 ni 500; todos deben ser 401
  // ----------------------------------------------------------------------------
  group('1. Acceso Anónimo a Endpoints Protegidos (Debe retornar 401)', () => {
    for (const ep of PROTECTED_ENDPOINTS) {
      const res = http.request(ep.method, `${BASE_URL}${ep.url}`, ep.body || null, {
        headers: { 'Content-Type': 'application/json' },
      });

      check(res, {
        [`${ep.name} bloqueado sin token (HTTP 401)`]: (r) => r.status === 401,
        [`${ep.name} mensaje de sesión requerido`]: (r) => {
          const body = r.json();
          return body.error && body.error.length > 0;
        },
      });
    }
  });

  // ----------------------------------------------------------------------------
  // TEST 2: TOKENS MANIPULADOS O INVÁLIDOS
  // Prueba con firma falsa (secreto incorrecto) y formato corrupto
  // ----------------------------------------------------------------------------
  group('2. Acceso con Token Forjado / Malformado (Debe retornar 401)', () => {
    // Token sintético para probar que firmas no autorizadas son rechazadas
    const fakeToken = ['eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9', 'eyJzdWIiOiIxIn0', 'fake_test_signature'].join('.');
    const malformedToken = 'BearerEstoNoEsUnTokenValido';

    for (const ep of PROTECTED_ENDPOINTS) {
      // Intento A: Firma forjada
      const resFake = http.request(ep.method, `${BASE_URL}${ep.url}`, ep.body || null, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${fakeToken}`,
        },
      });

      check(resFake, {
        [`${ep.name} rechaza firma forjada (HTTP 401)`]: (r) => r.status === 401,
      });

      // Intento B: Formato corrupto
      const resMalformed = http.request(ep.method, `${BASE_URL}${ep.url}`, ep.body || null, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${malformedToken}`,
        },
      });

      check(resMalformed, {
        [`${ep.name} rechaza token malformado (HTTP 401)`]: (r) => r.status === 401,
      });
    }
  });

  // ----------------------------------------------------------------------------
  // TEST 3: ENDPOINTS PÚBLICOS
  // Verificar que los catálogos, periodo y health check NO exigen token
  // ----------------------------------------------------------------------------
  group('3. Disponibilidad de Endpoints Públicos sin Autenticación', () => {
    for (const ep of PUBLIC_ENDPOINTS) {
      const res = http.request(ep.method, `${BASE_URL}${ep.url}`, null, {
        headers: { 'Content-Type': 'application/json' },
      });

      check(res, {
        [`${ep.name} es accesible públicamente (HTTP 200)`]: (r) => r.status === 200,
      });
    }
  });

  // ----------------------------------------------------------------------------
  // TEST 4: PREVENCIÓN DE IDOR / BOLA (Insecure Direct Object Reference)
  // Verificar que un endpoint protegido no acepte parámetros manipulados de otro usuario
  // ----------------------------------------------------------------------------
  group('4. Prevención de IDOR (El postulante_id solo se extrae del JWT)', () => {
    // Si un atacante envía ?postulante_id=999 o un header X-Postulante-ID: 999 sin token válido
    const resIdor = http.get(`${BASE_URL}/api/postulacion?postulante_id=999`, {
      headers: {
        'Content-Type': 'application/json',
        'X-User-ID': '999',
        'X-Postulante-ID': '999',
      },
    });

    check(resIdor, {
      'Consulta con ID forzado en query param sin token es rechazada (HTTP 401)': (r) => r.status === 401,
    });
  });
}

