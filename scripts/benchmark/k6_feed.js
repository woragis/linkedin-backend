import http from 'k6/http';
import { check, sleep } from 'k6';

// k6 run scripts/benchmark/k6_feed.js
// Requires: seed-demo + valid JWT in K6_TOKEN env

export const options = {
  vus: 20,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<500'],
  },
};

const BASE = __ENV.API_URL || 'http://127.0.0.1:8080';
const TOKEN = __ENV.K6_TOKEN || '';

export default function () {
  const headers = TOKEN ? { Authorization: `Bearer ${TOKEN}` } : {};
  const res = http.get(`${BASE}/v1/feed`, { headers });
  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(0.5);
}
