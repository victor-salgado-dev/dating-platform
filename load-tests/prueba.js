import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOTAL_USERS = parseInt(__ENV.USERS || '100', 10);
const MAX_PAGE = parseInt(__ENV.MAX_PAGE || '3', 10);
const DISCOVER_FILTER_MODE = __ENV.DISCOVER_FILTER_MODE || 'random';
const PASSWORD = __ENV.PASSWORD || 'DemoPass123!';
const rateLimited = new Counter('rate_limited_429');
const smoke = __ENV.MODE === 'smoke';

const SAMPLE_INTERESTS = ['viajar', 'cine', 'fitness', 'fotografia', 'cocina', 'musica', 'lectura', 'videojuegos', 'tecnologia', 'arte', 'senderismo'];
const SAMPLE_LANGUAGES = ['it', 'zh_cmn', 'de', 'ar', 'es', 'ja', 'en', 'ru'];
const SAMPLE_RELATIONSHIP_GOALS = ['casual', 'long_term', 'friendship', 'marriage', 'not_sure'];
const STABLE_DISCOVER_FILTERS = [
  { minAge: 18, maxAge: 30, gender: 'female', interests: ['viajar', 'cine'] },
  { minAge: 20, maxAge: 34, gender: 'male', language: 'es' },
  { minAge: 22, maxAge: 38, gender: 'non_binary', relationshipGoal: 'friendship' },
  { minAge: 25, maxAge: 40, gender: '', interests: ['fitness'] },
  { minAge: 18, maxAge: 35, gender: 'female', language: 'en' },
  { minAge: 27, maxAge: 45, gender: 'male', relationshipGoal: 'long_term' },
  { minAge: 19, maxAge: 33, gender: '', interests: ['musica', 'lectura'] },
  { minAge: 24, maxAge: 42, gender: 'female', language: 'de' },
  { minAge: 21, maxAge: 36, gender: 'male', relationshipGoal: 'casual' },
  { minAge: 28, maxAge: 48, gender: 'non_binary', interests: ['senderismo'] },
  { minAge: 18, maxAge: 32, gender: '', language: 'it' },
  { minAge: 23, maxAge: 39, gender: 'female', relationshipGoal: 'marriage' },
];

export const options = {
  stages: smoke
    ? [{ duration: '10s', target: 5 }, { duration: '20s', target: 5 }]
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
    'http_req_duration{endpoint:online}': ['p(95)<300'],
    'http_req_duration{endpoint:qm_batch}': ['p(95)<500'],
  },
};

export function setup() {
  const sessions = [];
  for (let i = 1; i <= TOTAL_USERS; i++) {
    const res = http.post(
      BASE_URL + '/api/v1/auth/login',
      JSON.stringify({ email: 'user' + i + '@datingdemo.com', password: PASSWORD }),
      { headers: { 'Content-Type': 'application/json' }, tags: { endpoint: 'login' } }
    );
    if (res.status === 200 && res.cookies.session_id) sessions.push(res.cookies.session_id[0].value);
  }
  if (sessions.length === 0) throw new Error('No se pudo iniciar sesión con ningún usuario de carga.');
  return { sessions };
}

function randInt(min, max) { return Math.floor(Math.random() * (max - min + 1)) + min; }
function pick(list) { return list[randInt(0, list.length - 1)]; }
function parseJson(res) { try { return res.json(); } catch (_) { return null; } }
function checkRateLimit(res) { if (res.status === 429) rateLimited.add(1); }

function buildDiscoverUrl(page) {
  const stableFilters = DISCOVER_FILTER_MODE === 'stable';
  const filterProfile = stableFilters ? STABLE_DISCOVER_FILTERS[(__VU - 1) % STABLE_DISCOVER_FILTERS.length] : null;
  const minAge = filterProfile ? filterProfile.minAge : randInt(18, 32);
  const maxAge = filterProfile ? filterProfile.maxAge : minAge + randInt(6, 20);
  const sort = pick(['recent', 'popular', 'age_asc', 'age_desc']);
  const gender = filterProfile ? filterProfile.gender : pick(['female', 'male', 'non_binary', '']);
  const filterType = filterProfile ? null : randInt(0, 2);
  let query = 'page=' + page + '&page_size=24&min_age=' + minAge + '&max_age=' + maxAge + '&sort=' + sort;
  if (gender) query += '&gender=' + encodeURIComponent(gender);

  if (filterProfile && filterProfile.interests) {
    query += '&interests=' + encodeURIComponent(filterProfile.interests.join(','));
  } else if (filterProfile && filterProfile.language) {
    query += '&language=' + encodeURIComponent(filterProfile.language);
  } else if (filterProfile && filterProfile.relationshipGoal) {
    query += '&relationship_goals=' + encodeURIComponent(filterProfile.relationshipGoal);
  } else if (filterType === 0) {
    const interestCount = randInt(1, 3);
    const pickedInterests = [];
    while (pickedInterests.length < interestCount) {
      const interest = pick(SAMPLE_INTERESTS);
      if (!pickedInterests.includes(interest)) pickedInterests.push(interest);
    }
    query += '&interests=' + encodeURIComponent(pickedInterests.join(','));
  } else if (filterType === 1) {
    query += '&language=' + encodeURIComponent(pick(SAMPLE_LANGUAGES));
  } else {
    query += '&relationship_goals=' + encodeURIComponent(pick(SAMPLE_RELATIONSHIP_GOALS));
  }
  return BASE_URL + '/api/v1/search/profiles?' + query;
}

export default function (data) {
  const sessionId = data.sessions[(__VU - 1) % data.sessions.length];
  const authHeaders = (tag) => ({
    headers: { Cookie: 'session_id=' + sessionId, 'Content-Type': 'application/json' },
    tags: { endpoint: tag },
  });
  const actions = randInt(2, 4);

  for (let i = 0; i < actions; i++) {
    const roll = Math.random() * 100;
    if (roll < 30) {
      const res = http.get(buildDiscoverUrl(randInt(1, MAX_PAGE)), authHeaders('discover'));
      checkRateLimit(res);
      const body = parseJson(res);
      check(res, { 'discover status 200': (r) => r.status === 200 }, { endpoint: 'discover' });
      check(body, { 'discover valid items': (b) => b !== null && Array.isArray(b.items) }, { endpoint: 'discover' });
    } else if (roll < 50) {
      const res = http.get(BASE_URL + '/api/v1/search/profiles?page=1&page_size=50', authHeaders('qm_batch'));
      checkRateLimit(res);
      check(res, { 'quick match batch status 200': (r) => r.status === 200 }, { endpoint: 'qm_batch' });
    } else if (roll < 70) {
      const res = http.get(BASE_URL + '/api/v1/search/recommended?page=1&page_size=24', authHeaders('recommended'));
      checkRateLimit(res);
      check(res, { 'recommended status 200': (r) => r.status === 200 }, { endpoint: 'recommended' });
    } else if (roll < 80) {
      const res = http.get(BASE_URL + '/api/v1/search/online-now?page=1&page_size=24', authHeaders('online'));
      checkRateLimit(res);
      check(res, { 'online status 200': (r) => r.status === 200 }, { endpoint: 'online' });
    } else if (roll < 90) {
      const res = http.get(BASE_URL + '/api/v1/search/popular?page=1&page_size=24', authHeaders('popular'));
      checkRateLimit(res);
      check(res, { 'popular status 200': (r) => r.status === 200 }, { endpoint: 'popular' });
    } else {
      const res = http.get(BASE_URL + '/api/v1/search/new-members?page=1&page_size=24', authHeaders('new'));
      checkRateLimit(res);
      check(res, { 'new members status 200': (r) => r.status === 200 }, { endpoint: 'new' });
    }
    sleep(0.5 + Math.random());
  }
  sleep(1 + Math.random() * 1.5);
}
