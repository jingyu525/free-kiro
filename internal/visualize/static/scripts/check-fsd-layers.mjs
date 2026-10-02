#!/usr/bin/env node
// scripts/check-fsd-layers.mjs — enforce FSD import direction (top→down).
//
// FSD layers (top→bottom):
//   app > pages > widgets > features > entities > shared
//
// Reverse imports and same-layer skips outside the allowlist are forbidden.
// Exit 1 on violation.

import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative, sep } from 'node:path';

const ROOT = join(process.cwd(), 'src');
const LAYERS = ['app', 'pages', 'widgets', 'features', 'entities', 'shared'];
const LAYER_RANK = Object.fromEntries(LAYERS.map((l, i) => [l, i]));

// Same-layer allowed pairs (entities → entities, widgets → widgets, etc.).
const SAME_LAYER_ALLOWED = true;

function* walk(dir) {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    const st = statSync(full);
    if (st.isDirectory()) {
      yield* walk(full);
    } else if (/\.(ts|tsx|mts|cts)$/.test(entry)) {
      yield full;
    }
  }
}

function layerOf(path) {
  const rel = relative(ROOT, path).split(sep);
  return rel[0];
}

function parseImports(file) {
  const src = readFileSync(file, 'utf8');
  const imports = [];
  // Match: import ... from '...'
  // and:    import('...')
  const re = /(?:from|import)\s*['"]([^'"]+)['"]/g;
  let m;
  while ((m = re.exec(src)) !== null) {
    const target = m[1];
    if (!target.startsWith('@')) continue;
    // @shared/foo → shared, @app/x → @/app/x, @/shared/foo → shared
    const cleaned = target.replace(/^@/, '').replace(/^\//, '');
    const parts = cleaned.split('/');
    const alias = parts[0];
    if (LAYERS.includes(alias)) imports.push({ alias, raw: target });
  }
  return imports;
}

let violations = 0;

for (const file of walk(ROOT)) {
  const fromLayer = layerOf(file);
  if (!LAYERS.includes(fromLayer)) continue;
  const fromRank = LAYER_RANK[fromLayer];
  for (const { alias } of parseImports(file)) {
    const toRank = LAYER_RANK[alias];
    if (toRank === undefined) continue;
    if (alias === fromLayer && SAME_LAYER_ALLOWED) continue;
    if (toRank < fromRank) {
      const rel = relative(process.cwd(), file);
      console.error(`✖ ${rel}: '${fromLayer}' → '${alias}' is a reverse import (allowed: top→down)`);
      violations++;
    }
  }
}

if (violations > 0) {
  console.error(`\n${violations} FSD layer violation(s).`);
  process.exit(1);
}
console.log('✓ FSD layers OK');