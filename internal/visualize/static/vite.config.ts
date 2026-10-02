// Vite config for free-kiro serve dashboard.
// Output goes to dist/assets/index-<hash>.{js,css} for //go:embed.
//
// Usage:
//   pnpm dev        # dev server on :5173
//   pnpm build      # production bundle to dist/
//   pnpm preview    # preview the built bundle on :4173

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { copyFileSync } from 'node:fs';
import { fileURLToPath, URL } from 'node:url';

export default defineConfig(({ mode: viteMode }) => {
  const isProd = viteMode === 'production';
  return {
    root: fileURLToPath(new URL('./src', import.meta.url)),
    // Use absolute root base so Vite emits `/assets/...` URLs (matching
    // the handleStatic route at /assets/<rel>); output goes to
    // dist/assets/ subdirectory which handleStatic falls back to.
    base: '/',
    publicDir: fileURLToPath(new URL('./public', import.meta.url)),
    plugins: [
      react(),
      {
        name: 'sync-static-index-html',
        closeBundle() {
          // Mirror dist/index.html to ../index.html so //go:embed picks up
          // the freshly-hashed <script>/<link> tags. Runs after every build.
          copyFileSync(
            fileURLToPath(new URL('./dist/index.html', import.meta.url)),
            fileURLToPath(new URL('./index.html', import.meta.url)),
          );
        },
      },
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '@app': fileURLToPath(new URL('./src/app', import.meta.url)),
        '@pages': fileURLToPath(new URL('./src/pages', import.meta.url)),
        '@widgets': fileURLToPath(new URL('./src/widgets', import.meta.url)),
        '@features': fileURLToPath(new URL('./src/features', import.meta.url)),
        '@entities': fileURLToPath(new URL('./src/entities', import.meta.url)),
        '@shared': fileURLToPath(new URL('./src/shared', import.meta.url)),
      },
    },
    server: {
      port: 5173,
      strictPort: false,
      host: '127.0.0.1',
    },
    preview: {
      port: 4173,
      host: '127.0.0.1',
    },
    build: {
      outDir: fileURLToPath(new URL('./dist', import.meta.url)),
      emptyOutDir: true,
      target: 'es2022',
      minify: 'esbuild',
      sourcemap: !isProd,
      cssMinify: 'esbuild',
      rollupOptions: {
        output: {
          // Output under dist/assets/ so //go:embed static/dist/ picks them up
          // and handleStatic at /assets/<rel> matches via the fallback lookup.
          entryFileNames: 'assets/[name]-[hash].js',
          chunkFileNames: 'assets/[name]-[hash].js',
          assetFileNames: 'assets/[name]-[hash][extname]',
          manualChunks: (id: string): string | undefined => {
            if (id.includes('node_modules/react/') || id.includes('node_modules/react-dom/')) {
              return 'react-vendor';
            }
            if (id.includes('node_modules/@tanstack/')) {
              return 'query-vendor';
            }
            return undefined;
          },
        },
      },
    },
  };
});