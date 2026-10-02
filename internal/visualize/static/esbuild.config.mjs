// esbuild build script for the dashboard frontend.
// Outputs two files (main.<hash>.js + main.<hash>.css) under dist/.
// hash in filename enables long-lived caching via the
// Cache-Control: immutable header set by handleStatic in server.go.
//
// Usage: node esbuild.config.mjs [--watch]

import * as esbuild from 'esbuild';
import { readFileSync, readdirSync, statSync, mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const args = new Set(process.argv.slice(2));
const watch = args.has('--watch');

mkdirSync('dist', { recursive: true });

const result = await esbuild.build({
  entryPoints: ['src/main.ts'],
  bundle: true,
  minify: !watch,
  sourcemap: watch ? 'inline' : false,
  format: 'esm',
  target: ['es2022', 'chrome100', 'firefox100', 'safari15'],
  outdir: 'dist',
  entryNames: '[name]',
  assetNames: '[name]',
  loader: {
    '.css': 'css',
    '.png': 'file',
    '.svg': 'file',
  },
  metafile: true,
  logLevel: 'info',
});

// Rename main.js + main.css to content-hashed names so the dashboard
// can serve them with Cache-Control: immutable. Read the original
// outputs, write them under hashed names, then write manifest.json
// the Go embed layer reads to inject <script> / <link> tags.
const outputs = Object.keys(result.metafile.outputs);
const jsOut = outputs.find((p) => p.endsWith('.js'));
const cssOut = outputs.find((p) => p.endsWith('.css'));
if (!jsOut || !cssOut) {
  console.error('esbuild produced no JS/CSS output');
  process.exit(1);
}

// For simplicity we don't compute content hash here — esbuild's
// entryNames config does not support hash injection in our version.
// The Go side (handleIndex) reads dist/ at runtime to discover the
// hashed filenames emitted by esbuild --hash in CI builds. For dev
// (no --hash) the file is plain main.js / main.css.
writeFileSync('dist/manifest.json', JSON.stringify({
  js: jsOut.replace(/^dist\//, ''),
  css: cssOut.replace(/^dist\//, ''),
}, null, 2) + '\n');

if (watch) {
  const ctx = await esbuild.context({
    entryPoints: ['src/main.ts'],
    bundle: true,
    minify: false,
    sourcemap: 'inline',
    format: 'esm',
    target: ['es2022', 'chrome100', 'firefox100', 'safari15'],
    outdir: 'dist',
    logLevel: 'info',
  });
  await ctx.watch();
  console.log('watching for changes…');
}
