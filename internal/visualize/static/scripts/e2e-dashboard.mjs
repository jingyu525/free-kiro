#!/usr/bin/env node
// scripts/e2e-dashboard.mjs — minimal HTTP-level end-to-end check.
//
// Verifies the dashboard wiring without a real browser (which would need
// Playwright + a Chromium install). Covers the dashboard-frontend-components
// AC-31 six scenarios:
//
//   1) GET / returns 200 with cache headers
//   2) GET /api/health returns status=ok
//   3) GET /api/summary returns JSON with .generated_at + .specs[]
//   4) GET /api/spec/<name> returns JSON
//   5) GET /api/spec/<name>/tasks returns array (possibly empty)
//   6) GET /api/spec/<name>/drift returns array (possibly empty)
//
// SSE refresh propagation is checked in the sibling smoke-dashboard.sh.

import { request } from 'node:http';

const PORT = process.env.E2E_PORT ? Number(process.env.E2E_PORT) : 7379;
const HOST = process.env.E2E_HOST ?? '127.0.0.1';

function get(path) {
  return new Promise((resolve, reject) => {
    const req = request(
      { host: HOST, port: PORT, path, method: 'GET' },
      (res) => {
        let body = '';
        res.on('data', (chunk) => (body += chunk));
        res.on('end', () => resolve({ status: res.statusCode ?? 0, body, headers: res.headers }));
      },
    );
    req.on('error', reject);
    req.end();
  });
}

let failures = 0;
const fail = (msg) => {
  failures++;
  console.error(`✖ ${msg}`);
};
const pass = (msg) => console.log(`✓ ${msg}`);

// 1) GET /
{
  const r = await get('/');
  if (r.status !== 200) fail(`/ expected 200, got ${r.status}`);
  else if (!String(r.headers['cache-control'] ?? '').includes('no-cache'))
    fail(`/ missing no-cache header (got: ${r.headers['cache-control']})`);
  else pass('/ returns 200 + no-cache');
}

// 2) /api/health
{
  const r = await get('/api/health');
  if (r.status !== 200) fail(`/api/health expected 200, got ${r.status}`);
  else {
    try {
      const j = JSON.parse(r.body);
      if (j.status !== 'ok') fail(`/api/health status: ${j.status}`);
      else pass('/api/health status=ok');
    } catch (e) {
      fail(`/api/health invalid JSON: ${(e).message}`);
    }
  }
}

// 3) /api/summary
let firstSpecName = null;
{
  const r = await get('/api/summary');
  if (r.status !== 200) fail(`/api/summary expected 200, got ${r.status}`);
  else {
    try {
      const j = JSON.parse(r.body);
      if (!j.generated_at) fail('/api/summary missing generated_at');
      else if (!Array.isArray(j.specs)) fail('/api/summary missing .specs[]');
      else if (j.specs.length === 0) fail('/api/summary .specs[] is empty (no test subject)');
      else {
        firstSpecName = j.specs[0]?.meta?.name;
        pass(`/api/summary has generated_at + ${j.specs.length} specs`);
      }
    } catch (e) {
      fail(`/api/summary invalid JSON: ${(e).message}`);
    }
  }
}

// 4) /api/spec/<name>
if (firstSpecName) {
  const r = await get(`/api/spec/${encodeURIComponent(firstSpecName)}`);
  if (r.status !== 200) fail(`/api/spec/${firstSpecName} expected 200, got ${r.status}`);
  else pass(`/api/spec/${firstSpecName} returns 200`);
}

// 5) /api/spec/<name>/tasks
if (firstSpecName) {
  const r = await get(`/api/spec/${encodeURIComponent(firstSpecName)}/tasks`);
  if (r.status !== 200) fail(`/api/spec/${firstSpecName}/tasks expected 200, got ${r.status}`);
  else {
    const j = JSON.parse(r.body);
    if (!j || !Array.isArray(j.waves)) fail(`/api/spec/${firstSpecName}/tasks missing waves[]`);
    else pass(`/api/spec/${firstSpecName}/tasks returns {spec, waves[]}`);
  }
}

// 6) /api/spec/<name>/drift
if (firstSpecName) {
  const r = await get(`/api/spec/${encodeURIComponent(firstSpecName)}/drift`);
  if (r.status !== 200) fail(`/api/spec/${firstSpecName}/drift expected 200, got ${r.status}`);
  else {
    const j = JSON.parse(r.body);
    if (!j || !Array.isArray(j.signals)) fail(`/api/spec/${firstSpecName}/drift missing signals[]`);
    else pass(`/api/spec/${firstSpecName}/drift returns {spec, signals[]}`);
  }
}

if (failures > 0) {
  console.error(`\n${failures} E2E failure(s).`);
  process.exit(1);
}
console.log('\n✓ E2E dashboard passed (6 scenarios)');