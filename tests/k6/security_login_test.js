import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';

// ==============================================================================
// SRPD BIESS — Suite de Pruebas de Seguridad en Autenticación con k6
// Pista D (Datos & QA) — Responsable: Gino Bermúdez (@ginobermudez76)
// Issue: #23 — Rate Limiting, Prevención de Enumeración y Timing Attacks
// ==============================================================================

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Métricas personalizadas para análisis de temporización (Timing Attacks)
const trendExistente = new Trend('tiempo_usuario_existente_ms', true);
const trendInexistente = new Trend('tiempo_usuario_inexistente_ms', true);
const blockedCounter = new Counter('intentos_bloqueados_429');

export const options = {
  scenarios: {
    // 1. Escenario de Seguridad y Timing (Secuencial para análisis preciso)
    seguridad_y_timing: {
      executor: 'per-vu-iterations',
      vus: 1,
      iterations: 1,
      maxDuration: '1m',
    },
  },
  thresholds: {
    // Umbrales de aceptación para OWASP A07
    'checks': ['rate==1.0'], // Todas las aserciones funcionales de seguridad deben pasar
  },
};

const HEADERS = {
  'Content-Type': 'application/json',
};

export default function () {
  console.log(`[k6] Iniciando suite de seguridad contra: ${BASE_URL}`);

  // ----------------------------------------------------------------------------
  // PRUEBA 1: UNICIDAD DE MENSAJES (PREVENCIÓN DE ENUMERACIÓN DE USUARIOS)
  // ----------------------------------------------------------------------------
  group('1. Prevención de Enumeración de Usuarios (Mensaje Unificado S6)', () => {
    // Caso A: Cédula que existe en base de datos pero con password errónea
    const payloadExistente = JSON.stringify({
      cedula: '1710034065', // Juan Carlos Pérez (semilla)
      password: 'ClaveInvalida123*',
    });

    // Usamos cédula aleatoria que no existe para no colisionar con intentos
    const cedulaInexistente = '0999999999';
    const payloadInexistente = JSON.stringify({
      cedula: cedulaInexistente,
      password: 'ClaveInvalida123*',
    });

    const resExistente = http.post(`${BASE_URL}/api/auth/login`, payloadExistente, { headers: HEADERS });
    const resInexistente = http.post(`${BASE_URL}/api/auth/login`, payloadInexistente, { headers: HEADERS });

    const jsonExistente = resExistente.json();
    const jsonInexistente = resInexistente.json();

    const mensajeEsperado = 'sus credenciales de ingreso son incorrectas';

    check(resExistente, {
      'Usuario existente con clave errónea retorna 401': (r) => r.status === 401,
      'Usuario existente retorna mensaje genérico exacto': () => jsonExistente.error === mensajeEsperado,
    });

    check(resInexistente, {
      'Usuario inexistente retorna 401': (r) => r.status === 401,
      'Usuario inexistente retorna mensaje genérico exacto': () => jsonInexistente.error === mensajeEsperado,
    });

    check(null, {
      'Ambos mensajes de error son estrictamente IDÉNTICOS': () => jsonExistente.error === jsonInexistente.error,
    });
  });

  sleep(1);

  // ----------------------------------------------------------------------------
  // PRUEBA 2: RATE LIMITING Y BLOQUEO TEMPORAL (EF-11 / DECISIÓN S4)
  // Regla: 3 intentos fallidos consecutivos -> bloqueo de 10 minutos (429)
  // ----------------------------------------------------------------------------
  group('2. Rate Limiting y Bloqueo de 10 min al 3er Intento (EF-11)', () => {
    // Usamos una cédula específica para probar el contador de intentos en Redis
    const cedulaTest = '0912345675'; // Ana Lucía Mora (semilla)
    const payloadFallo = JSON.stringify({
      cedula: cedulaTest,
      password: 'PasswordIncorrecto999#',
    });

    console.log(`[k6] Ejecutando 4 intentos consecutivos para cédula: ${cedulaTest}`);

    // Intento 1 -> 401
    const res1 = http.post(`${BASE_URL}/api/auth/login`, payloadFallo, { headers: HEADERS });
    check(res1, {
      'Intento 1 retorna 401': (r) => r.status === 401,
    });

    // Intento 2 -> 401
    const res2 = http.post(`${BASE_URL}/api/auth/login`, payloadFallo, { headers: HEADERS });
    check(res2, {
      'Intento 2 retorna 401': (r) => r.status === 401,
    });

    // Intento 3 -> 429 Bloqueado (Tercer fallo consecutivo)
    const res3 = http.post(`${BASE_URL}/api/auth/login`, payloadFallo, { headers: HEADERS });
    const json3 = res3.json();
    if (res3.status === 429) blockedCounter.add(1);

    check(res3, {
      'Intento 3 retorna 429 Too Many Requests': (r) => r.status === 429,
      'Intento 3 informa bloqueo de 10 minutos en mensaje': () => 
        json3.error && json3.error.includes('10 minutos'),
    });

    // Intento 4 -> Debe permanecer 429 (bloqueo vigente)
    const res4 = http.post(`${BASE_URL}/api/auth/login`, payloadFallo, { headers: HEADERS });
    if (res4.status === 429) blockedCounter.add(1);

    check(res4, {
      'Intento 4 permanece en 429 durante el bloqueo': (r) => r.status === 429,
    });
  });

  sleep(1);

  // ----------------------------------------------------------------------------
  // PRUEBA 3: ANÁLISIS DE TIEMPOS DE RESPUESTA (TIMING ATTACK / DISCREPANCIA BCRYPT)
  // ----------------------------------------------------------------------------
  group('3. Análisis de Discrepancia de Tiempo (Timing Attack Analysis)', () => {
    console.log('[k6] Midiendo latencias comparativas entre usuario existente vs inexistente...');

    const muestras = 5;

    for (let i = 0; i < muestras; i++) {
      // Cédula que existe en base de datos (con bcrypt real ejecutado si no está bloqueada)
      // Generamos variaciones para medir el costo computacional de bcrypt
      const resExistente = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
        cedula: '0102030400', // Pedro Salazar (semilla)
        password: 'ClaveErroneaParaTiming#1',
      }), { headers: HEADERS });

      trendExistente.add(resExistente.timings.duration);

      // Cédula inexistente (no llega a bcrypt si no se mitiga con dummy hash)
      const resInexistente = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
        cedula: `09${Math.floor(10000000 + Math.random() * 89999999)}`,
        password: 'ClaveErroneaParaTiming#1',
      }), { headers: HEADERS });

      trendInexistente.add(resInexistente.timings.duration);

      sleep(0.2);
    }
  });
}

