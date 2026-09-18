import { defineConfig } from 'vite';
import { fileURLToPath } from 'node:url';

export default defineConfig({
  root: fileURLToPath(new URL('.', import.meta.url)),
  server: { port: 5173, strictPort: true },
  preview: { port: 4173, strictPort: true },
  optimizeDeps: { exclude: ['@ffmpeg/ffmpeg'] },
  build: { target: 'es2022', outDir: 'dist', emptyOutDir: true },
});
