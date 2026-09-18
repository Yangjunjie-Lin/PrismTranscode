import { mkdir, copyFile } from 'node:fs/promises';
const root = new URL('../', import.meta.url);
const dest = new URL('online/public/engine/', root);
await mkdir(dest, { recursive: true });
for (const name of ['ffmpeg-core.js', 'ffmpeg-core.wasm']) {
  await copyFile(new URL(`node_modules/@ffmpeg/core/dist/esm/${name}`, root), new URL(name, dest));
}
await copyFile(new URL('node_modules/@ffmpeg/core/package.json', root), new URL('package.json', dest));
await copyFile(new URL('THIRD_PARTY_NOTICES.md', root), new URL('online/public/THIRD_PARTY_NOTICES.txt', root));
await copyFile(new URL('LICENSE', root), new URL('online/public/LICENSE.txt', root));
await copyFile(new URL('licenses/FFmpeg-GPL-2.0.txt', root), new URL('online/public/FFmpeg-GPL-2.0.txt', root));
console.log('Pinned FFmpeg WebAssembly engine copied for same-origin hosting.');
