import { LocalizedError, message } from './i18n.js';
export const MAX_FILE = 128 * 1024 * 1024;
export const MAX_QUEUE = 384 * 1024 * 1024;
export const TARGETS = [
  { id: 'mp3', kind: 'audio', ext: 'mp3', mime: 'audio/mpeg', encoder: 'libmp3lame' },
  { id: 'wav', kind: 'audio', ext: 'wav', mime: 'audio/wav', encoder: 'pcm_s16le' },
  { id: 'flac', kind: 'audio', ext: 'flac', mime: 'audio/flac', encoder: 'flac' },
  { id: 'm4a', kind: 'audio', ext: 'm4a', mime: 'audio/mp4', encoder: 'aac' },
  { id: 'ogg', kind: 'audio', ext: 'ogg', mime: 'audio/ogg', encoder: 'libvorbis' },
  { id: 'opus', kind: 'audio', ext: 'opus', mime: 'audio/ogg', encoder: 'opus' },
  { id: 'mp4', kind: 'video', ext: 'mp4', mime: 'video/mp4', encoder: 'libx264' },
  { id: 'webm', kind: 'video', ext: 'webm', mime: 'video/webm', encoder: 'libvpx' },
  { id: 'gif', kind: 'image', ext: 'gif', mime: 'image/gif', encoder: 'gif' },
  { id: 'png', kind: 'image', ext: 'png', mime: 'image/png', encoder: 'png' },
  { id: 'jpg', kind: 'image', ext: 'jpg', mime: 'image/jpeg', encoder: 'mjpeg' },
  { id: 'webp', kind: 'image', ext: 'webp', mime: 'image/webp', encoder: 'libwebp' },
  { id: 'srt', kind: 'subtitle', ext: 'srt', mime: 'text/plain', encoder: 'srt' },
  { id: 'vtt', kind: 'subtitle', ext: 'vtt', mime: 'text/vtt', encoder: 'webvtt' },
];
export function target(id) {
  const value = TARGETS.find(t => t.id === id);
  if (!value) throw new LocalizedError('error.target');
  return value;
}
export function validateOptions(options) {
  target(options.target);
  if (!['balanced', 'high', 'small'].includes(options.quality)) throw new LocalizedError('error.quality');
  for (const key of ['start', 'duration']) {
    if (!Number.isFinite(options[key]) || options[key] < 0 || options[key] > 86400) throw new LocalizedError('error.time');
  }
  if (![0, 480, 720, 1080].includes(options.resolution)) throw new LocalizedError('error.resolution');
  return options;
}
export function outputName(name, id, sequence = 1) {
  const stem = name.replace(/\.[^.]+$/, '').replace(/[\x00-\x1f<>:"/\\|?*]/g, '_').slice(0, 120) || 'converted';
  return `${stem}_prism_${sequence}.${target(id).ext}`;
}
export function inputExtension(name) {
  return /\.([a-z0-9]{1,8})$/i.exec(name)?.[1]?.toLowerCase() || 'media';
}
export function fileError(file, queueBytes = 0) {
  if (!file.size) return message('error.emptyFile');
  if (inputExtension(file.name) === 'ncm') return message('error.ncm');
  if (file.size > MAX_FILE) return message('error.fileSize');
  if (file.size + queueBytes > MAX_QUEUE) return message('error.queueSize');
  return '';
}
export function buildArgs(input, output, options) {
  const o = validateOptions(options), t = target(o.target);
  const args = ['-hide_banner', '-loglevel', 'info', '-nostdin', '-y', '-threads', '1', '-filter_threads', '1', '-protocol_whitelist', 'file,pipe', '-i', input];
  if (o.start) args.push('-ss', String(o.start));
  if (o.duration && t.kind !== 'image') args.push('-t', String(o.duration));
  if (t.kind === 'audio') {
    args.push('-map', '0:a:0', '-vn', '-c:a', t.encoder);
    if (['mp3', 'm4a'].includes(t.id)) args.push('-b:a', { high: '320k', balanced: '192k', small: '96k' }[o.quality]);
    if (t.id === 'opus') args.push('-strict', '-2', '-ar', '48000', '-ac', '2', '-b:a', { high: '192k', balanced: '128k', small: '64k' }[o.quality]);
    if (t.id === 'ogg') args.push('-q:a', { high: '8', balanced: '5', small: '2' }[o.quality]);
  } else if (t.kind === 'video') {
    args.push('-map', '0:v:0', '-map', '0:a:0?', '-sn', '-c:v', t.encoder, '-crf', { high: '20', balanced: '26', small: '32' }[o.quality]);
    if (t.id === 'mp4') args.push('-preset', 'ultrafast', '-pix_fmt', 'yuv420p', '-c:a', 'aac', '-b:a', '160k', '-movflags', '+faststart');
    else args.push('-b:v', { high: '2M', balanced: '1M', small: '500k' }[o.quality], '-deadline', 'realtime', '-cpu-used', '5', '-c:a', 'libvorbis', '-q:a', '5');
    const scale = o.resolution ? `scale=w='min(${Math.round(o.resolution * 16 / 9)},iw)':h='min(${o.resolution},ih)':force_original_aspect_ratio=decrease:force_divisible_by=2` : 'scale=trunc(iw/2)*2:trunc(ih/2)*2';
    args.push('-vf', scale);
  } else if (t.kind === 'image') {
    args.push('-map', '0:v:0', '-an');
    if (t.id === 'gif') args.push('-t', String(Math.min(o.duration || 10, 10)), '-vf', "fps=12,scale='min(480,iw)':-1:flags=lanczos");
    else {
      args.push('-frames:v', '1', '-c:v', t.encoder);
      if (t.id === 'webp') args.push('-lossless', '1');
      if (t.id === 'jpg') args.push('-q:v', { high: '2', balanced: '4', small: '8' }[o.quality]);
    }
  } else args.push('-map', '0:s:0', '-c:s', t.encoder);
  return [...args, '-threads', '1', output];
}
export function rejectHDR(probe, id) {
  const t = target(id);
  if (!['video', 'image'].includes(t.kind)) return;
  if ((probe.streams || []).some(s => ['smpte2084', 'arib-std-b67'].includes(s.color_transfer))) {
    throw new LocalizedError('error.hdr');
  }
}
