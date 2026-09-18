// Keep the upstream source, build recipes, dependency sources and their licenses
// available next to the redistributed, unmodified GPL WebAssembly binary.
import { mkdir, writeFile, readFile } from 'node:fs/promises';
import { createHash } from 'node:crypto';
const sources = [
  ['ffmpegwasm/ffmpeg.wasm', 'v12.15'], ['FFmpeg/FFmpeg', 'n5.1.4'],
  ['ffmpegwasm/x264', '4-cores'], ['ffmpegwasm/x265', '3.4'],
  ['ffmpegwasm/libvpx', 'v1.13.1'], ['ffmpegwasm/lame', 'master'],
  ['ffmpegwasm/Ogg', 'v1.3.4'], ['ffmpegwasm/theora', 'v1.1.1'],
  ['ffmpegwasm/opus', 'v1.3.1'], ['ffmpegwasm/vorbis', 'v1.3.3'],
  ['ffmpegwasm/zlib', 'v1.2.11'], ['ffmpegwasm/libwebp', 'v1.3.2'],
  ['ffmpegwasm/freetype2', 'VER-2-10-4'], ['fribidi/fribidi', 'v1.0.9'],
  ['harfbuzz/harfbuzz', '5.2.0'], ['libass/libass', '0.15.0'],
  ['sekrit-twc/zimg', 'release-3.0.5'], ['libsdl-org/SDL', 'release-2.24.2'],
  ['google/googletest', '703bd9caab50b139428cea1aaff9974ebee5742e'],
  ['emscripten-core/emscripten', '3.1.40'],
];
const dest = new URL('../artifacts/wasm-sources/', import.meta.url);
await mkdir(dest, { recursive: true });
const manifest = { core: '@ffmpeg/core@0.12.10', wrapper: '@ffmpeg/ffmpeg@0.12.15', generated: new Date().toISOString(), sources: [] };
async function download(url) {
  const response = await fetch(url, { signal: AbortSignal.timeout(180000) });
  if (!response.ok) throw new Error(`${url}: HTTP ${response.status}`);
  return Buffer.from(await response.arrayBuffer());
}
for (let i = 0; i < sources.length; i += 3) {
  await Promise.all(sources.slice(i, i + 3).map(async ([repo, ref]) => {
    const url = `https://codeload.github.com/${repo}/tar.gz/${ref}`;
    const name = `${repo.replace('/', '-')}-${ref}.tar.gz`;
    const data = await download(url);
    await writeFile(new URL(name, dest), data);
    manifest.sources.push({ repo, ref, url, archive: name, sha256: createHash('sha256').update(data).digest('hex') });
    console.log(name, data.length);
  }));
}
manifest.sources.sort((a, b) => a.repo.localeCompare(b.repo));
for (const name of ['ffmpeg-core.js', 'ffmpeg-core.wasm']) {
  const data = await readFile(new URL(`../node_modules/@ffmpeg/core/dist/esm/${name}`, import.meta.url));
  (manifest.binaries ||= []).push({ name, sha256: createHash('sha256').update(data).digest('hex') });
}
await writeFile(new URL('MANIFEST.json', dest), JSON.stringify(manifest, null, 2));
await writeFile(new URL('BUILD.md', dest), `# FFmpeg WebAssembly source bundle\n\nUpstream: https://github.com/ffmpegwasm/ffmpeg.wasm/tree/v12.15\n\nThis bundle contains the unmodified upstream source, Dockerfile, build scripts, FFmpeg n5.1.4 and sources for libraries selected in that Dockerfile, plus SDL2 (Emscripten port). Each source archive preserves its upstream licenses. The npm core binary is redistributed without modification; hashes are recorded in MANIFEST.json.\n\nTo build: extract ffmpegwasm-ffmpeg.wasm-v12.15.tar.gz, install Docker/Buildx, and follow its Makefile's build-st target (Emscripten 3.1.40, FFMPEG_ST=yes). Use the bundled dependency archives as the Dockerfile source inputs for offline source availability; the Emscripten SDK and build-system packages are general-purpose tooling obtained separately.\n\nThe upstream recipes use branch names for x264 and LAME; the downloaded snapshots are archived and hashed, but byte-for-byte reproduction of the npm binary is not certified. Do not misrepresent this source collection as an independent license, patent or security audit.\n`);
const gpl = await download('https://raw.githubusercontent.com/FFmpeg/FFmpeg/n5.1.4/COPYING.GPLv2');
await writeFile(new URL('../licenses/FFmpeg-GPL-2.0.txt', import.meta.url), gpl);
await writeFile(new URL('COPYING.GPLv2', dest), gpl);
console.log('Source bundle ready; package artifacts/wasm-sources with tar.');
