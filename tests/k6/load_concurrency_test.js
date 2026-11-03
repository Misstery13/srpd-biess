import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';

// ==============================================================================
// SRPD BIESS — Suite de Concurrencia y Carga Masiva con k6
// Pista D (Datos, Rendimiento & QA) — Responsable: Gino Bermúdez (@ginobermudez76)
// Issue: #25 — Estrés Transaccional, pgxpool (15 conns) y Caché Redis
// ==============================================================================

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Configurar códigos HTTP esperados para que 401 (rechazo auth) y 429 (rate limiting / bloqueo)
// no sean considerados fallos de red en http_req_failed por k6 (Issue #25)
http.setResponseCallback(http.expectedStatuses(200, 401, 429));

// Métricas personalizadas para medir el impacto de la memoria y la base de datos
const trendCatalogos = new Trend('latencia_catalogos_redis_ms', true);
const trendPeriodo = new Trend('latencia_periodo_redis_ms', true);
const trendLogin = new Trend('latencia_login_concurrente_ms', true);
const peticionesCacheExitosas = new Counter('peticiones_cache_exitosas');

export const options = {
  // Rampa de carga progresiva simulando tráfico concurrente de postulantes
  stages: [
    { duration: '5s', target: 20 },  // Rampa de subida a 20 VUs simultáneos
    { duration: '15s', target: 50 }, // Carga sostenida a 50 VUs simultáneos
    { duration: '5s', target: 70 },  // Pico de estrés (spike) a 70 VUs simultáneos
    { duration: '5s', target: 0 },   // Rampa de enfriamiento a 0 VUs
  ],
  thresholds: {
    // Umbrales de calidad para alto rendimiento
    'http_req_duration': ['p(95)<300'], // 95% de peticiones totales bajo 300 ms
    'latencia_catalogos_redis_ms': ['p(95)<80'], // Lecturas cacheadas en Redis deben ser ultra rápidas (< 80ms)
    'http_req_failed': ['rate<0.05'], // Tasa de errores no esperados menor al 5%
  },
};

const HEADERS = {
  'Content-Type': 'application/json',
};

// Cédulas de prueba para rotación concurrente
const CEDULAS_POOL = [
  '1710034065', // Juan Pérez
  '0912345675', // Ana Mora
  '0102030400', // Pedro Salazar
  '1754321980', // Carlos Gómez
  '0102030405',
  '1712345678',
  '0911223344',
  '1799887766',
];

export default function loadConcurrencyTest() {
  const vuId = __VU;
  const iterId = __ITER;

  // ----------------------------------------------------------------------------
  // FLUJO 1: LECTURA MASIVA DE CATÁLOGOS Y PERÍODO (EVALUACIÓN DEL CACHÉ REDIS)
  // ----------------------------------------------------------------------------
  // 1. Catálogo de Países (Cache-Aside)
  const resPaises = http.get(`${BASE_URL}/api/catalogos/paises`, { headers: HEADERS });
  trendCatalogos.add(resPaises.timings.duration);
  if (resPaises.status === 200) peticionesCacheExitosas.add(1);

  check(resPaises, {
    'GET /api/catalogos/paises responde 200': (r) => r.status === 200,
  });

  // 2. Período de Postulación (Cache-Aside)
  const resPeriodo = http.get(`${BASE_URL}/api/periodo`, { headers: HEADERS });
  trendPeriodo.add(resPeriodo.timings.duration);
  if (resPeriodo.status === 200) peticionesCacheExitosas.add(1);

  check(resPeriodo, {
    'GET /api/periodo responde 200': (r) => r.status === 200,
  });

  // 3. Cascada geográfica (Provincias de Ecuador - ID: 52)
  const resProvincias = http.get(`${BASE_URL}/api/catalogos/paises/52/provincias`, { headers: HEADERS });
  trendCatalogos.add(resProvincias.timings.duration);
  if (resProvincias.status === 200) peticionesCacheExitosas.add(1);

  check(resProvincias, {
    'GET /api/catalogos/paises/52/provincias responde 200': (r) => r.status === 200,
  });

  // ----------------------------------------------------------------------------
  // FLUJO 2: CONCURRENCIA TRANSACCIONAL Y RATE LIMITING (EVALUACIÓN DE POSTGRES & REDIS)
  // ----------------------------------------------------------------------------
  // Selección determinista de cédula según el VU para probar concurrencia multi-usuario
  const cedulaSeleccionada = CEDULAS_POOL[(vuId + iterId) % CEDULAS_POOL.length];

  const payloadLogin = JSON.stringify({
    cedula: cedulaSeleccionada,
    password: `SimulatedEntry_VU_${vuId}_${iterId}`,
  });

  const resLogin = http.post(`${BASE_URL}/api/auth/login`, payloadLogin, { headers: HEADERS });
  trendLogin.add(resLogin.timings.duration);

  // Bajo alta concurrencia, el endpoint debe responder 401 (credenciales erróneas) o 429 (bloqueo por 3 intentos)
  // Jamás debe responder 500 (colapso de servidor ni pánico)
  check(resLogin, {
    'POST /api/auth/login responde 401 o 429 bajo estrés': (r) => r.status === 401 || r.status === 429,
    'Servidor no genera errores 500 por contención': (r) => r.status !== 500,
  });

  // Jitter determinista entre 50ms y 140ms para distribuir peticiones concurrentes
  sleep(0.05 + ((iterId % 10) * 0.01));
}

