export const LANGUAGE_KEY = 'prismtranscode.language';
// Keep both languages together so new UI messages cannot silently go untranslated.
export const messages = {
  'page.title': ['流光转码 PrismTranscode — 在线媒体工作台', 'PrismTranscode — Private online media workbench'],
  'page.description': ['在浏览器内批量转换音频、短视频、图片和字幕，媒体文件不上传服务器。', 'Batch-convert audio, short videos, images and subtitles in your browser. Your media stays on your device.'],
  'brand.name': ['流光转码', 'PrismTranscode'],
  'brand.subtitle': ['PrismTranscode', 'LOCAL MEDIA WORKBENCH'],
  'brand.home': ['流光转码首页', 'PrismTranscode home'],
  'nav.label': ['主导航', 'Main navigation'],
  'nav.workspace': ['在线工作台', 'Workbench'],
  'nav.help': ['使用说明', 'How it works'],
  'nav.desktop': ['下载桌面版 ↓', 'Get desktop ↓'],
  'language.label': ['界面语言 / Language', 'Language / 界面语言'],
  'language.auto': ['跟随浏览器', 'Browser default'],
  'hero.eyebrow': ['你的文件，你的设备。', 'YOUR FILES. YOUR DEVICE.'],
  'hero.line1': ['格式自由，', 'New formats.'],
  'hero.line2': ['原作', 'Your files, '],
  'hero.accent': ['留在你手里。', 'still yours.'],
  'hero.description': ['从一段声音，到每一帧画面。\n打开浏览器，让文件成为你需要的格式。', 'From a single sound to every frame.\nTurn your files into the formats you need, right in your browser.'],
  'hero.noInstall': ['↳ 无需安装', '↳ No installation'],
  'hero.private': ['◉ 媒体不上传', '◉ No media uploads'],
  'hero.openSource': ['⌘ 开源工作台', '⌘ Open source'],
  'art.title': ['一个输入，更多可能。', 'ONE INPUT. NEW POSSIBILITIES.'],
  'art.engine': ['由 FFMPEG 驱动', 'POWERED BY FFMPEG'],
  'art.local': ['全部在浏览器内完成 ↗', '100% IN-BROWSER ↗'],
  'workspace.kicker': ['媒体工作台', 'THE WORKBENCH'],
  'workspace.title': ['你的转换工作台', 'Your conversion workspace'],
  'queue.label': ['转换队列', 'Conversion queue'],
  'queue.add': ['添加文件', 'Add files'],
  'queue.count': ['{count} 个文件', '{count} files'],
  'queue.countOne': ['{count} 个文件', '{count} file'],
  'queue.originals': ['原文件不修改，结果由你下载保存。', 'Originals stay unchanged. Download your results to keep them.'],
  'queue.summary': ['{done} / {total} 已完成 · 输入 {size}', '{done} / {total} complete · {size} input'],
  'queue.clear': ['清空队列', 'Clear queue'],
  'queue.empty': ['还没有待转换的文件', 'No files in your queue yet'],
  'queue.emptyHint': ['每个文件都能选择自己的目标格式', 'Choose a different output format for each file.'],
  'queue.local': ['◉ 仅在当前浏览器内处理', '◉ Processed only in this browser'],
  'queue.export': ['导出转换记录 ↗', 'Export results ↗'],
  'drop.title': ['把文件拖到这里', 'Drop your files here'],
  'drop.or': ['或', 'or'],
  'drop.browse': ['点击选择文件', 'browse your device'],
  'drop.limit': ['音频、短视频、图片、文本字幕 · 单文件 ≤ 128 MiB', 'Audio, short videos, images, text subtitles · ≤128 MiB per file'],
  'settings.label': ['输出设置', 'Output settings'],
  'settings.title': ['设置输出', 'Set your output'],
  'settings.defaults': ['默认参数', 'Default options'],
  'settings.categories': ['格式类别', 'Format category'],
  'kind.audio': ['音频', 'Audio'], 'kind.video': ['视频', 'Video'],
  'kind.image': ['图片', 'Images'], 'kind.subtitle': ['字幕', 'Subtitles'],
  'settings.target': ['目标格式', 'Output format'],
  'settings.quality': ['质量偏好', 'Quality preference'],
  'quality.balanced': ['均衡 · 日常推荐', 'Balanced · recommended'],
  'quality.high': ['高质量 · 较大文件', 'High quality · larger files'],
  'quality.small': ['小体积 · 更多压缩', 'Smaller files · more compression'],
  'settings.advanced': ['裁剪与视频尺寸', 'Trim & video size'],
  'settings.start': ['开始时间 / 秒', 'Start / seconds'],
  'settings.duration': ['时长 / 秒', 'Duration / seconds'],
  'settings.durationHint': ['时长为 0 表示直到结束；GIF 最长 10 秒。', '0 means until the end. GIFs are limited to 10 seconds.'],
  'settings.resolution': ['视频尺寸上限', 'Maximum video size'],
  'settings.originalSize': ['原始尺寸', 'Original size'],
  'settings.apply': ['将参数应用到未完成任务', 'Apply to unfinished files'],
  'settings.truth': ['保留真实，不虚构提升。', 'Preserve what is real.'],
  'settings.lossless': ['无损格式不会恢复原本丢失的细节。', 'Lossless formats cannot restore detail already lost.'],
  'settings.startButton': ['开始转换', 'Start converting'],
  'settings.stop': ['停止当前批次', 'Stop this batch'],
  'settings.footnote': ['第一次转换时下载引擎。大文件、NCM 和 GPU 编码请使用桌面版。', 'The engine downloads on first use. For large files, NCM or GPU encoding, use the desktop app.'],
  'about.kicker': ['为本地处理而设计', 'MADE TO STAY LOCAL'],
  'about.title': ['在线使用，\n不等于上传云端。', 'Online does not mean\nuploading your files.'],
  'about.description': ['网站与引擎通过网络加载，媒体内容只进入浏览器内存。没有账号，没有媒体上传接口，没有追踪脚本。', 'The site and engine load over the internet. Your media stays in browser memory. No account, media upload endpoint or tracking scripts.'],
  'about.guide': ['阅读完整使用指南 ↗', 'Read the full user guide ↗'],
  'faq.files.title': ['在线版适合哪些文件？', 'Which files work online?'],
  'faq.files.body': ['音频、短视频、静态图片和文本字幕。支持 14 个输出预设，具体可用项由引擎检测。建议使用最新版桌面浏览器；端到端实测基线为 Chromium，移动设备可能因内存不足而失败。', 'Audio, short videos, still images and text subtitles. There are 14 output presets, subject to engine capabilities. Use a current desktop browser; Chromium is the tested end-to-end baseline. Mobile devices may run out of memory.'],
  'faq.desktop.title': ['和桌面版有什么区别？', 'How is it different from desktop?'],
  'faq.desktop.body': ['桌面版提供 38 个输出预设、NCM、本地路径输入、多任务并行及硬件编码选项。在线版单文件最多 128 MiB，输入队列总计最多 384 MiB / 50 个文件，逐个处理，无 GPU 加速；HDR 转图片或视频会明确拒绝。', 'Desktop offers 38 output presets, NCM, direct local paths, parallel jobs and hardware encoding options. Online limits are 128 MiB per file and 384 MiB / 50 files per queue, processed sequentially without GPU acceleration. HDR-to-image/video conversion is rejected.'],
  'faq.storage.title': ['我的文件会保存多久？', 'How long are my files kept?'],
  'faq.storage.body': ['文件只保留在当前标签页内存中，刷新或关闭页面即丢失。请逐个下载结果；清空或移除任务会释放对应结果。手动语言偏好单独保存在此浏览器中，不保存媒体。', 'Files remain in this tab’s memory and are lost on refresh or close. Download each result; clearing or removing a job releases it. Only your manual language preference is saved in this browser, not your media.'],
  'faq.limits.title': ['哪些内容不在本版保证范围内？', 'What is not guaranteed?'],
  'faq.limits.body': ['不保证复杂多轨、附件和所有元数据完整迁移，不提供 HDR 色调映射或不可信媒体安全认证。视频只保留第一条画面和音轨。WebM 使用 VP8 + Vorbis；Opus 使用实验性编码器。FFmpeg WebAssembly 含 GPL 组件，详见许可。', 'No guarantee of preserving all tracks, attachments or metadata; no HDR tone mapping or security certification for untrusted media. Video keeps the first video and audio streams. WebM uses VP8 + Vorbis; Opus uses an experimental encoder. FFmpeg WebAssembly includes GPL components; see the notices.'],
  'footer.beta': ['开源媒体工具 · 社区预发布', 'Open-source media tools · community beta'],
  'footer.licenses': ['第三方许可', 'Third-party notices'],
  'footer.issues': ['问题反馈 ↗', 'Report an issue ↗'],
  'footer.tagline': ['让格式少一点阻碍。', 'Fewer barriers. More possibilities.'],
  'engine.idle': ['○ 引擎按需加载 · 约 32 MB', '○ Engine loads on demand · about 32 MB'],
  'engine.loading': ['◌ 正在下载并初始化引擎 · 约 32 MB', '◌ Loading the engine · about 32 MB'],
  'engine.ready': ['● 本地引擎就绪 · {count} 个输出预设', '● Local engine ready · {count} output presets'],
  'engine.failed': ['○ 引擎加载失败 · 可重试', '○ Engine unavailable · retry to reload'],
  'engine.stopped': ['○ 已停止 · 下次转换重新初始化', '○ Stopped · engine will reload on next conversion'],
  'engine.finished': ['● 本批次结束 · 引擎内存已释放', '● Batch finished · engine memory released'],
  'engine.unavailable': [' · 引擎不可用', ' · unavailable'],
  'engine.noFormats': ['当前引擎没有此类别的编码器', 'No encoders are available for this category.'],
  'state.ready': ['等待转换', 'Ready to convert'], 'state.running': ['正在转换', 'Converting'],
  'state.completed': ['转换完成', 'Completed'], 'state.failed': ['转换失败', 'Failed'],
  'state.cancelled': ['已停止，可重试', 'Stopped · ready to retry'],
  'job.progress': ['{name} 转换进度', 'Conversion progress for {name}'],
  'job.target': ['{name} 的输出格式', 'Output format for {name}'],
  'job.download': ['下载 ↓', 'Download ↓'], 'job.downloadLabel': ['下载 {name}', 'Download {name}'],
  'job.remove': ['移除', 'Remove'], 'job.removeLabel': ['移除 {name}', 'Remove {name}'],
  'job.verification': ['FFprobe 输出结构检查 + SHA-256（非全文件解码或画质认证）', 'FFprobe output structure check + SHA-256 (not full decoding or quality certification)'],
  'error.target': ['不支持的输出格式', 'Unsupported output format.'],
  'error.quality': ['质量参数无效', 'Invalid quality option.'],
  'error.time': ['起点与时长需为 0–86400 秒', 'Start and duration must be between 0 and 86400 seconds.'],
  'error.resolution': ['分辨率参数无效', 'Invalid resolution option.'],
  'error.emptyFile': ['文件为空', 'The file is empty.'],
  'error.ncm': ['NCM 请使用桌面版；在线版不提供 NCM 解码', 'Use the desktop app for NCM; online NCM decoding is not available.'],
  'error.fileSize': ['单文件上限为 128 MiB，请使用桌面版处理大文件', 'The per-file limit is 128 MiB. Use the desktop app for larger files.'],
  'error.queueSize': ['队列上限为 384 MiB，请先移除部分文件', 'The queue limit is 384 MiB. Remove some files first.'],
  'error.queueCount': ['队列最多 50 个文件', 'The queue can hold up to 50 files.'],
  'error.hdr': ['检测到 HDR；在线版不执行色调映射，请使用经过验证的桌面工作流', 'HDR detected. Online tone mapping is not available; use a validated desktop workflow.'],
  'error.stopped': ['已停止', 'Stopped.'],
  'error.encoders': ['引擎编码器检测失败', 'Could not detect the engine’s encoders.'],
  'error.probe': ['无法识别媒体结构：文件可能损坏或格式不受支持', 'Could not identify the media. The file may be damaged or unsupported.'],
  'error.encoder': ['此输出编码器在在线引擎中不可用，请改用桌面版', 'This encoder is unavailable online. Use the desktop app.'],
  'error.timeout': ['此任务超过 5 分钟，已释放引擎；请缩短片段或使用桌面版', 'This job exceeded 5 minutes and the engine was released. Use a shorter clip or the desktop app.'],
  'error.convert': ['转换失败，请检查输入是否包含所需音轨、画面或字幕', 'Conversion failed. Check that the input contains the required audio, video or subtitle stream.'],
  'error.output': ['输出校验失败：没有可识别的媒体流', 'Output verification failed: no recognizable media streams.'],
  'error.emptyOutput': ['转换输出为空', 'The conversion produced an empty file.'],
  'error.engine': ['引擎加载或执行失败。请检查网络、浏览器兼容性与可用内存，再点击开始重试。', 'The engine could not load or run. Check your connection, browser compatibility and available memory, then retry.'],
  'error.unknown': ['转换失败；请重试或使用桌面版', 'Conversion failed. Retry or use the desktop app.'],
  'error.browser': ['当前浏览器不支持所需的安全上下文、WebAssembly 或 Worker，请使用最新版桌面浏览器通过 HTTPS 打开。', 'A secure context, WebAssembly and Workers are required. Open this site over HTTPS in a current desktop browser.'],
  'format.mp3.label': ['MP3', 'MP3'], 'format.mp3.hint': ['通用兼容 · 音频转换与视频声音提取', 'Widely compatible · convert audio or extract video sound'],
  'format.wav.label': ['WAV', 'WAV'], 'format.wav.hint': ['未压缩 PCM · 16-bit 音频', 'Uncompressed PCM · 16-bit audio'],
  'format.flac.label': ['FLAC', 'FLAC'], 'format.flac.hint': ['无损压缩 · 不恢复源文件已丢失的信息', 'Lossless compression · cannot restore previously lost detail'],
  'format.m4a.label': ['M4A / AAC', 'M4A / AAC'], 'format.m4a.hint': ['小巧通用 · 适合移动设备', 'Compact and compatible · suited to mobile devices'],
  'format.ogg.label': ['OGG / Vorbis', 'OGG / Vorbis'], 'format.ogg.hint': ['开放格式 · Vorbis 编码', 'Open format · Vorbis encoding'],
  'format.opus.label': ['Opus · 实验性', 'Opus · experimental'], 'format.opus.hint': ['FFmpeg 实验性编码器 · 48 kHz 立体声；高要求请用桌面版 libopus', 'Experimental FFmpeg encoder · 48 kHz stereo; use desktop libopus for demanding work'],
  'format.mp4.label': ['MP4 / H.264', 'MP4 / H.264'], 'format.mp4.hint': ['H.264 + AAC · 适合短视频，浏览器内编码较慢', 'H.264 + AAC · for short clips; in-browser encoding is slower'],
  'format.webm.label': ['WebM / VP8', 'WebM / VP8'], 'format.webm.hint': ['VP8 + Vorbis · 浏览器兼容视频；VP9 请使用桌面版', 'VP8 + Vorbis · browser-friendly video; use desktop for VP9'],
  'format.gif.label': ['GIF 动图', 'Animated GIF'], 'format.gif.hint': ['最长 10 秒 · 480 像素宽 · 12 帧/秒', 'Up to 10 seconds · 480 pixels wide · 12 fps'],
  'format.png.label': ['PNG', 'PNG'], 'format.png.hint': ['无损静态图片 · 视频仅提取第一帧', 'Lossless still image · extracts only the first video frame'],
  'format.jpg.label': ['JPEG', 'JPEG'], 'format.jpg.hint': ['通用静态图片 · 不保留透明通道', 'Widely supported still image · no transparency'],
  'format.webp.label': ['WebP 无损', 'Lossless WebP'], 'format.webp.hint': ['无损静态图片 · 保留透明通道', 'Lossless still image · preserves transparency'],
  'format.srt.label': ['SRT', 'SRT'], 'format.srt.hint': ['文本字幕 · 不支持图像字幕 OCR', 'Text subtitles · no OCR for image-based subtitles'],
  'format.vtt.label': ['WebVTT', 'WebVTT'], 'format.vtt.hint': ['网页文本字幕 · 不保证复杂样式一致', 'Web text subtitles · complex styling may not be preserved'],
};

let language = 'en';
export function resolveLanguage(preference, languages = []) {
  if (preference === 'zh' || preference === 'en') return preference;
  for (const tag of languages) {
    if (typeof tag !== 'string') continue;
    const primary = tag.toLowerCase().split(/[-_]/)[0];
    if (primary === 'zh' || primary === 'en') return primary;
  }
  return 'en';
}
export function getLanguage() { return language; }
export function setLanguage(value) { language = value === 'zh' ? 'zh' : 'en'; }
export function t(key, values = {}, locale = language) {
  const pair = messages[key];
  if (!pair) throw new Error(`Missing translation: ${key}`);
  return pair[locale === 'zh' ? 0 : 1].replace(/\{(\w+)\}/g, (match, name) => String(values[name] ?? match));
}
export function message(key, values = {}) { return { key, values }; }
export class LocalizedError extends Error {
  constructor(key, values = {}) { super(t(key, values)); this.key = key; this.values = values; }
}
export function messageText(value) {
  if (!value) return '';
  return value.key ? t(value.key, value.values) : typeof value === 'string' ? value : t('error.unknown');
}
export function applyTranslations(root = document) {
  root.documentElement.lang = language === 'zh' ? 'zh-CN' : 'en';
  root.title = t('page.title');
  root.querySelector('meta[name="description"]').content = t('page.description');
  root.querySelectorAll('[data-i18n]').forEach(el => { el.textContent = t(el.dataset.i18n); });
  root.querySelectorAll('[data-i18n-aria]').forEach(el => { el.setAttribute('aria-label', t(el.dataset.i18nAria)); });
}
