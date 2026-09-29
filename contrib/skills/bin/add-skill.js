#!/usr/bin/env node
// @jingyu525/free-kiro-skill
//
// Thin wrapper that downloads install.sh from the repo and exec's it
// with the user-supplied env. Keeps `npx @jingyu525/free-kiro-skill`
// working without depending on the upstream `add-skill` registry.

const { execFileSync } = require('child_process');
const https = require('https');
const path = require('path');
const fs = require('fs');

const REPO = process.env.FREE_KIRO_SKILL_REPO || 'jingyu525/free-kiro';
const RAW_URL = `https://raw.githubusercontent.com/${REPO}/main/contrib/skills/install.sh`;

function fetch(url) {
  return new Promise((resolve, reject) => {
    https.get(url, (res) => {
      if (res.statusCode && res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return fetch(res.headers.location).then(resolve, reject);
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
      }
      let data = '';
      res.on('data', (c) => (data += c));
      res.on('end', () => resolve(data));
      res.on('error', reject);
    }).on('error', reject);
  });
}

(async () => {
  console.error(`fetching install.sh from ${RAW_URL}…`);
  const script = await fetch(RAW_URL);
  const tmp = path.join(require('os').tmpdir(), `free-kiro-install-${Date.now()}.sh`);
  fs.writeFileSync(tmp, script, { mode: 0o755 });
  try {
    // Forward all FREE_KIRO_SKILL_* env vars.
    const env = { ...process.env };
    if (!env.FREE_KIRO_SKILL_REPO) env.FREE_KIRO_SKILL_REPO = REPO;
    execFileSync('bash', [tmp], { stdio: 'inherit', env });
  } finally {
    try { fs.unlinkSync(tmp); } catch (_) {}
  }
})().catch((err) => {
  console.error('error:', err.message);
  process.exit(1);
});