export const MAX_FILE = 128 * 1024 * 1024;
export const MAX_QUEUE = 384 * 1024 * 1024;
export const TARGETS = [
  { id: 'mp3', label: 'MP3', kind: 'audio', ext: 'mp3', mime: 'audio/mpeg', encoder: 'libmp3lame', hint: '通用兼容 · 音频转换与视频声音提取' },
  { id: 'wav', label: 'WAV', kind: 'audio', ext: 'wav', mime: 'audio/wav', encoder: 'pcm_s16le', hint: '未压缩 PCM · 16-bit 音频' },
  { id: 'flac', label: 'FLAC', kind: 'audio', ext: 'flac', mime: 'audio/flac', encoder: 'flac', hint: '无损压缩 · 不恢复源文件已丢失的信息' },
  { id: 'm4a', label: 'M4A / AAC', kind: 'audio', ext: 'm4a', mime: 'audio/mp4', encoder: 'aac', hint: '小巧通用 · 适合移动设备' },
  { id: 'ogg', label: 'OGG / Vorbis', kind: 'audio', ext: 'ogg', mime: 'audio/ogg', encoder: 'libvorbis', hint: '开放格式 · Vorbis 编码' },
  { id: 'opus', label: 'Opus · 实验性', kind: 'audio', ext: 'opus', mime: 'audio/ogg', encoder: 'opus', hint: 'FFmpeg 实验性编码器 · 48 kHz 立体声；高要求请用桌面版 libopus' },
  { id: 'mp4', label: 'MP4 / H.264', kind: 'video', ext: 'mp4', mime: 'video/mp4', encoder: 'libx264', hint: 'H.264 + AAC · 适合短视频，浏览器内编码较慢' },
  { id: 'webm', label: 'WebM / VP8', kind: 'video', ext: 'webm', mime: 'video/webm', encoder: 'libvpx', hint: 'VP8 + Vorbis · 浏览器兼容视频；VP9 请使用桌面版' },
  { id: 'gif', label: 'GIF 动图', kind: 'image', ext: 'gif', mime: 'image/gif', encoder: 'gif', hint: '最长 10 秒 · 480 像素宽 · 12 帧/秒' },
  { id: 'png', label: 'PNG', kind: 'image', ext: 'png', mime: 'image/png', encoder: 'png', hint: '无损静态图片 · 视频仅提取第一帧' },
  { id: 'jpg', label: 'JPEG', kind: 'image', ext: 'jpg', mime: 'image/jpeg', encoder: 'mjpeg', hint: '通用静态图片 · 不保留透明通道' },
  { id: 'webp', label: 'WebP 无损', kind: 'image', ext: 'webp', mime: 'image/webp', encoder: 'libwebp', hint: '无损静态图片 · 保留透明通道' },
  { id: 'srt', label: 'SRT', kind: 'subtitle', ext: 'srt', mime: 'text/plain', encoder: 'srt', hint: '文本字幕 · 不支持图像字幕 OCR' },
  { id: 'vtt', label: 'WebVTT', kind: 'subtitle', ext: 'vtt', mime: 'text/vtt', encoder: 'webvtt', hint: '网页文本字幕 · 不保证复杂样式一致' },
];
export function target(id) {
  const value = TARGETS.find(t => t.id === id);
  if (!value) throw new Error('不支持的输出格式');
  return value;
}
export function validateOptions(options) {
  target(options.target);
  if (!['balanced', 'high', 'small'].includes(options.quality)) throw new Error('质量参数无效');
  for (const key of ['start', 'duration']) {
    if (!Number.isFinite(options[key]) || options[key] < 0 || options[key] > 86400) throw new Error('起点与时长需为 0–86400 秒');
  }
  if (![0, 480, 720, 1080].includes(options.resolution)) throw new Error('分辨率参数无效');
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
  if (!file.size) return '文件为空';
  if (inputExtension(file.name) === 'ncm') return 'NCM 请使用桌面版；在线版不提供 NCM 解码';
  if (file.size > MAX_FILE) return '单文件上限为 128 MiB，请使用桌面版处理大文件';
  if (file.size + queueBytes > MAX_QUEUE) return '队列上限为 384 MiB，请先移除部分文件';
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
    throw new Error('检测到 HDR；在线版不执行色调映射，请使用经过验证的桌面工作流');
  }
}
