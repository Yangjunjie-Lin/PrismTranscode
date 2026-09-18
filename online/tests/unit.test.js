import test from 'node:test';
import assert from 'node:assert/strict';
import { TARGETS, buildArgs, fileError, inputExtension, outputName, validateOptions, rejectHDR, MAX_FILE, MAX_QUEUE } from '../convert.js';
const options = { target: 'mp3', quality: 'balanced', start: 0, duration: 0, resolution: 0 };
test('all 14 presets produce argument arrays with a fixed output', () => {
  assert.equal(TARGETS.length, 14);
  for (const t of TARGETS) { const args = buildArgs('input.media', `output.${t.ext}`, { ...options, target: t.id }); assert.equal(args.at(-1), `output.${t.ext}`); assert.ok(args.includes('-protocol_whitelist')); assert.ok(args.includes(t.encoder) || t.id === 'gif'); }
});
test('audio selects an actual audio stream', () => assert.ok(buildArgs('i', 'o', options).includes('0:a:0')));
test('rejects unknown formats, nonfinite numbers and invalid resolutions', () => {
  for (const o of [{ target: '-y' }, { start: NaN }, { duration: Infinity }, { start: -1 }, { quality: 'invalid' }, { resolution: 900 }]) assert.throws(() => validateOptions({ ...options, ...o }));
});
test('GIF duration is bounded and video resize does not upscale', () => {
  const gif = buildArgs('i', 'o', { ...options, target: 'gif', duration: 900 }); assert.equal(gif[gif.indexOf('-t') + 1], '10');
  const video = buildArgs('i', 'o', { ...options, target: 'mp4', resolution: 720 }); assert.ok(video.some(s => s.includes('min(720,ih)')));
});
test('file and queue limits, empty files and NCM are rejected', () => {
  assert.ok(fileError({ name: 'x.mp4', size: MAX_FILE + 1 })); assert.ok(fileError({ name: 'x.mp4', size: 2 }, MAX_QUEUE - 1)); assert.ok(fileError({ name: 'x', size: 0 })); assert.ok(fileError({ name: 'x.NCM', size: 4 })); assert.equal(fileError({ name: 'x.wav', size: 100 }), '');
});
test('filenames cannot become commands or paths', () => { assert.equal(inputExtension('x;rm -rf /'), 'media'); assert.equal(outputName('../你好.wav', 'mp3', 3), '.._你好_prism_3.mp3'); assert.throws(() => outputName('x', 'invalid')); });
test('HDR visual conversion is explicitly rejected but audio is permitted', () => { const info = { streams: [{ color_transfer: 'smpte2084' }] }; assert.throws(() => rejectHDR(info, 'mp4')); assert.throws(() => rejectHDR(info, 'png')); assert.doesNotThrow(() => rejectHDR(info, 'mp3')); });
