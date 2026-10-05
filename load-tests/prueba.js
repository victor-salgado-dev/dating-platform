import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOTAL_USERS = parseInt(__ENV.USERS || '100', 10);
const MAX_PAGE = parseInt(__ENV.MAX_PAGE || '3', 10);
const PASSWORD = __ENV.PASSWORD || 'DemoPass123!';

const rateLimited = new Counter('rate_limited_429');
const matchesFound = new Counter('matches_found');
const smoke = __ENV.MODE === 'smoke';

export const options = {
  stages: smoke
    ? [ { duration: '10s', target: 5 }, { duration: '20s', target: 5 } ]
    : [
        { duration: '30s', target: 20 },
        { duration: '1m', target: 50 },
        { duration: '1m', target: 100 },
        { duration: '1m', target: 150 },
        { duration: '1m', target: 250 },
        { duration: '1m', target: 400 },
        { duration: '30s', target: 0 },
      ],
  thresholds: {
    http_req_failed: [{ threshold: 'rate<0.05', abortOnFail: true, delayAbortEval: '30s' }],
    http_req_duration: ['p(95)<600'],
    'http_req_duration{endpoint:recommended}': ['p(95)<300'],
    'http_req_duration{endpoint:popular}': ['p(95)<300'],
    'http_req_duration{endpoint:discover}': ['p(95)<500'],
    'http_req_duration{endpoint:profile_detail}': ['p(95)<250'],
    'http_req_duration{endpoint:like}': ['p(95)<300'],
    'http_req_duration{endpoint:match_check}': ['p(95)<250'],
  },
};

export function setup() {
  const sessions = [];
  for (let i = 1; i <= TOTAL_USERS; i++) {
    const res = http.post(
      `${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({ email: `user${i}@datingdemo.com`, password: PASSWORD }),
      { headers: { 'Content-Type': 'application/json' }, tags: { endpoint: 'login' } }
    );
    if (res.status === 200 && res.cookies['session_id']) {
      sessions.push(res.cookies['session_id'][0].value);
    }
  }
  return { sessions };
}

function randInt(min, max) { return Math.floor(Math.random() * (max - min + 1)) + min; }
function pick(list) { return list[randInt(0, list.length - 1)]; }
function parseJson(res) { try { return res.json(); } catch (e) { return null; } }

const SAMPLE_INTERESTS = [
  'viajar', 'cine', 'fitness', 'fotografia', 'cocina', 'musica',
  'lectura', 'videojuegos', 'tecnologia', 'arte', 'senderismo'
];

function buildDiscoverUrl(page) {
  const minAge = randInt(18, 32);
  const maxAge = minAge + randInt(6, 20);
  const sort = pick(['recent', 'popular', 'age_asc', 'age_desc']);
  const gender = pick(['female', 'male', 'non_binary', '']);

  // Seleccionamos de 1 a 3 intereses (nunca superará el límite de 25)
  const interestCount = randInt(1, 3);
  const pickedInterests = [];
  for (let i = 0; i < interestCount; i++) {
    const item = pick(SAMPLE_INTERESTS);
    if (!pickedInterests.includes(item)) pickedInterests.push(item);
  }

  let query = `page=${page}&page_size=24&min_age=${minAge}&max_age=${maxAge}&sort=${sort}`;
  if (gender) query += `&gender=${gender}`;
  if (pickedInterests.length > 0) query += `&interests=${pickedInterests.join(',')}`;

  return `${BASE_URL}/api/v1/search/profiles?${query}`;
}

export default function (data) {
  const sessionId = data.sessions[__VU % data.sessions.length];
  const authHeaders = (tag) => ({
    headers: {
      Cookie: `session_id=${sessionId}`,
      'Content-Type': 'application/json',
    },
    tags: { endpoint: tag },
  });

  const actions = randInt(2, 4);

  for (let i = 0; i < actions; i++) {
    // 35% Discover, 25% QuickMatch con Likes/Matches, 20% Recommended, 10% Popular, 10% New
    const roll = Math.random() * 100;

    if (roll < 35) {
      // 1. DISCOVER (Búsqueda avanzada con filtros realistas)
      const page = randInt(1, MAX_PAGE);
      const url = buildDiscoverUrl(page);
      const res = http.get(url, authHeaders('discover'));
      if (res.status === 429) rateLimited.add(1);

      const body = parseJson(res);
      check(res, { 'discover status 200': (r) => r.status === 200 }, { endpoint: 'discover' });
      check(body, { 'discover valid items': (b) => b !== null && Array.isArray(b.items) }, { endpoint: 'discover' });

    } else if (roll < 60) {
      // 2. QUICK MATCH (Flujo de tarjetas, ver perfil, Like y Match status)
      // Carga la tanda de cartas (page_size 50)
      const batchRes = http.get(`${BASE_URL}/api/v1/search/profiles?page=1&page_size=50`, authHeaders('quickmatch_batch'));
      if (batchRes.status === 429) rateLimited.add(1);

      const batch = parseJson(batchRes);
      if (batch && Array.isArray(batch.items) && batch.items.length > 0) {
        // Tomamos una carta aleatoria de la tanda recibida
        const target = pick(batch.items);
        const targetId = target.profile_id;

        // A) Pide el perfil detallado completo (como hace useFullProfile)
        const profileRes = http.get(`${BASE_URL}/api/v1/profiles/${targetId}`, authHeaders('profile_detail'));
        check(profileRes, { 'profile detail 200': (r) => r.status === 200 }, { endpoint: 'profile_detail' });

        sleep(0.3);

        // B) Simula dar Like
        const likeRes = http.post(`${BASE_URL}/api/v1/likes/${targetId}`, null, authHeaders('like'));
        check(likeRes, { 'like 200 or 201': (r) => r.status === 200 || r.status === 201 }, { endpoint: 'like' });

        // C) Si dio like, el front comprueba si hubo match
        const matchRes = http.get(`${BASE_URL}/api/v1/matches/${targetId}`, authHeaders('match_check'));
        check(matchRes, { 'match check 200': (r) => r.status === 200 }, { endpoint: 'match_check' });

        const matchBody = parseJson(matchRes);
        if (matchBody && matchBody.matched === true) {
          matchesFound.add(1);
        }
      }

    } else if (roll < 80) {
      // 3. RECOMMENDED (Usa las preferencias de pareja que acabamos de poblar)
      const res = http.get(`${BASE_URL}/api/v1/search/recommended?page=1&page_size=24`, authHeaders('recommended'));
      check(res, { 'recommended status 200': (r) => r.status === 200 }, { endpoint: 'recommended' });

    } else if (roll < 90) {
      // 4. POPULAR
      const res = http.get(`${BASE_URL}/api/v1/search/popular?page=1&page_size=24`, authHeaders('popular'));
      check(res, { 'popular status 200': (r) => r.status === 200 }, { endpoint: 'popular' });

    } else {
      // 5. NEW MEMBERS
      const res = http.get(`${BASE_URL}/api/v1/search/new-members?page=1&page_size=24`, authHeaders('new'));
      check(res, { 'new members status 200': (r) => r.status === 200 }, { endpoint: 'new' });
    }

    sleep(0.5 + Math.random());
  }

  sleep(1 + Math.random() * 1.5);
}