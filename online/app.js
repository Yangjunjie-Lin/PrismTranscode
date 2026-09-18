import './style.css';
import './language.css';
import { FFmpeg } from '@ffmpeg/ffmpeg';
import { TARGETS, target, buildArgs, fileError, inputExtension, outputName, validateOptions, rejectHDR } from './convert.js';
import { LANGUAGE_KEY, resolveLanguage, getLanguage, setLanguage, t, message, LocalizedError, messageText, applyTranslations } from './i18n.js';

const $ = id => document.getElementById(id);
const jobs = [];
let engine = null, loading = null, busy = false, cancelled = false, current = null, counter = 0;
let kind = 'audio', encoders = null, logs = [];
let languagePreference = 'auto', notices = [], engineState = 'idle';
const browserSupported = 'WebAssembly' in window && 'Worker' in window && window.isSecureContext;
try { languagePreference = localStorage.getItem(LANGUAGE_KEY) || 'auto'; } catch { /* Storage may be blocked. */ }
if (!['zh', 'en'].includes(languagePreference)) languagePreference = 'auto';
setLanguage(resolveLanguage(languagePreference, navigator.languages?.length ? navigator.languages : [navigator.language]));
const bytes = size => size < 1024 * 1024 ? `${(size / 1024).toFixed(1)} KiB` : `${(size / 1024 / 1024).toFixed(1)} MiB`;
function renderNotice() {
  $('notice').textContent = notices.map(item => `${item.file ? `${item.file}: ` : ''}${messageText(item)}`).join('\n');
  $('notice').hidden = !notices.length;
}
function notice(items = []) { notices = (Array.isArray(items) ? items : [items]).filter(Boolean); renderNotice(); }
function renderEngine() { $('engineStatus').textContent = t(`engine.${engineState}`, { count: TARGETS.filter(available).length }); }
function setEngineState(state) { engineState = state; renderEngine(); }
function updateLanguage() {
  setLanguage(resolveLanguage(languagePreference, navigator.languages?.length ? navigator.languages : [navigator.language]));
  applyTranslations();
  $('languageSelect').value = languagePreference;
  $('guideLink').href = `https://github.com/Yangjunjie-Lin/PrismTranscode/blob/main/docs/USAGE.${getLanguage() === 'zh' ? 'zh-CN' : 'en'}.md`;
  $('licenseLink').href = getLanguage() === 'zh' ? '/THIRD_PARTY_NOTICES.txt' : '/THIRD_PARTY_NOTICES.en.txt';
  defaults(); render(); renderNotice(); renderEngine();
}
function options() { return validateOptions({ target: $('target').value, quality: $('quality').value, start: Number($('start').value), duration: Number($('duration').value), resolution: Number($('resolution').value) }); }
function available(t) { return !encoders || encoders.has(t.encoder); }
function fillTargets(select, category, selected) {
  select.replaceChildren();
  for (const format of TARGETS.filter(format => !category || format.kind === category)) {
    const option = new Option(`${t(`format.${format.id}.label`)}${available(format) ? '' : t('engine.unavailable')}`, format.id);
    option.disabled = !available(format); select.add(option);
  }
  if ([...select.options].some(o => o.value === selected && !o.disabled)) select.value = selected;
  else select.value = [...select.options].find(o => !o.disabled)?.value || '';
}
function hint() { $('targetHint').textContent = $('target').value ? t(`format.${$('target').value}.hint`) : t('engine.noFormats'); }
function defaults() { const selected = $('target').value; fillTargets($('target'), kind, selected); hint(); }
function revoke(job) { if (job.url) URL.revokeObjectURL(job.url); job.url = null; job.outputSize = 0; }
function render() {
  $('fileCount').textContent = t(jobs.length === 1 ? 'queue.countOne' : 'queue.count', { count: jobs.length });
  $('empty').hidden = jobs.length > 0;
  $('clear').disabled = busy || !jobs.length;
  $('apply').disabled = busy || !jobs.length;
  $('export').disabled = busy || !jobs.length;
  $('startButton').disabled = !browserSupported || busy || !jobs.some(j => j.state !== 'completed');
  $('startButton').hidden = busy; $('stopButton').hidden = !busy;
  $('dropzone').disabled = busy || !browserSupported;
  $('queueSummary').textContent = jobs.length ? t('queue.summary', { done: jobs.filter(j => j.state === 'completed').length, total: jobs.length, size: bytes(jobs.reduce((n, j) => n + j.file.size, 0)) }) : t('queue.originals');
  $('queue').replaceChildren(...jobs.map(job => {
    const row = document.createElement('li'); row.className = 'job'; row.dataset.state = job.state; row.dataset.id = job.id;
    const icon = document.createElement('span'); icon.className = 'file-icon'; icon.textContent = inputExtension(job.file.name).toUpperCase();
    const info = document.createElement('div');
    const name = document.createElement('span'); name.className = 'job-name'; name.textContent = job.file.name; name.title = job.file.name;
    const meta = document.createElement('span'); meta.className = 'job-meta'; meta.textContent = `${bytes(job.file.size)}${job.outputSize ? ` → ${bytes(job.outputSize)}` : ''}`;
    const progress = document.createElement('progress'); progress.max = 100; progress.value = job.progress || 0; progress.setAttribute('aria-label', t('job.progress', { name: job.file.name }));
    const status = document.createElement('div'); status.className = 'job-status'; status.textContent = messageText(job.error) || `${t(`state.${job.state}`)}${job.state === 'running' ? ` · ${Math.round(job.progress)}%` : ''}`;
    info.append(name, meta, progress, status);
    const controls = document.createElement('div'), select = document.createElement('select'); select.className = 'job-select';
    select.setAttribute('aria-label', t('job.target', { name: job.file.name })); fillTargets(select, null, job.options.target);
    // Do not silently replace a selected but unavailable format.
    select.value = job.options.target; select.disabled = busy || job.state === 'completed';
    select.addEventListener('change', () => { job.options.target = select.value; job.state = 'ready'; job.error = ''; render(); });
    const actions = document.createElement('div'); actions.className = 'job-actions';
    if (job.url) { const link = document.createElement('a'); link.href = job.url; link.download = job.output; link.textContent = t('job.download'); link.setAttribute('aria-label', t('job.downloadLabel', { name: job.output })); actions.append(link); }
    const remove = document.createElement('button'); remove.className = 'text-button'; remove.textContent = t('job.remove'); remove.disabled = busy;
    remove.setAttribute('aria-label', t('job.removeLabel', { name: job.file.name }));
    remove.addEventListener('click', () => { revoke(job); jobs.splice(jobs.indexOf(job), 1); render(); });
    actions.append(remove); controls.append(select, actions); row.append(icon, info, controls); return row;
  }));
}
function progressUI(job) {
  const row = document.querySelector(`[data-id="${job.id}"]`);
  if (!row) return;
  row.querySelector('progress').value = job.progress;
  row.querySelector('.job-status').textContent = `${t(`state.${job.state}`)} · ${Math.round(job.progress)}%`;
}
function addFiles(files) {
  if (busy || !browserSupported) return;
  const errors = [];
  let o; try { o = options(); } catch (error) { notice(error); return; }
  for (const file of files) {
    if (jobs.length >= 50) { errors.push(message('error.queueCount')); break; }
    if (jobs.some(j => j.file.name === file.name && j.file.size === file.size && j.file.lastModified === file.lastModified)) continue;
    const error = fileError(file, jobs.reduce((n, j) => n + j.file.size, 0));
    if (error) { errors.push({ ...error, file: file.name }); continue; }
    jobs.push({ id: ++counter, file, options: { ...o }, state: 'ready', progress: 0, error: '' });
  }
  notice(errors); render();
}
async function getEngine() {
  if (engine?.loaded) return engine;
  if (loading) return loading;
  const instance = new FFmpeg(); engine = instance;
  instance.on('log', ({ message }) => { logs.push(message); if (logs.length > 100) logs.shift(); if (new URLSearchParams(location.search).has('debug')) console.debug(message); });
  instance.on('progress', ({ progress }) => {
    if (current?.state !== 'running') return;
    current.progress = Math.max(current.progress, Math.min(93, Math.max(1, progress * 90)));
    progressUI(current);
  });
  loading = (async () => {
    setEngineState('loading');
    await instance.load({ coreURL: '/engine/ffmpeg-core.js', wasmURL: '/engine/ffmpeg-core.wasm' });
    if (cancelled) throw new LocalizedError('error.stopped');
    if (encoders) { setEngineState('ready'); return instance; }
    const encoderLines = [];
    const collect = ({ message }) => encoderLines.push(message);
    instance.on('log', collect);
    try {
      if (await instance.exec(['-hide_banner', '-encoders']) !== 0) throw new LocalizedError('error.encoders');
    } finally { instance.off('log', collect); }
    encoders = new Set(encoderLines.map(line => /^\s*[VAS][A-Z.]{5}\s+(\S+)/.exec(line)?.[1]).filter(Boolean));
    if (!encoders.size) throw new LocalizedError('error.encoders');
    defaults(); render();
    setEngineState('ready');
    return instance;
  })();
  try { return await loading; }
  catch (error) { instance.terminate(); if (engine === instance) engine = null; throw error; }
  finally { loading = null; }
}
async function probe(ff, path, jsonPath) {
  if (new URLSearchParams(location.search).has('debug')) console.debug('Probe start', path);
  const code = await ff.ffprobe(['-v', 'error', '-threads', '1', '-protocol_whitelist', 'file,pipe', '-show_streams', '-show_format', '-of', 'json', path, '-o', jsonPath], 30000);
  // core 0.12.10's ffprobe binding ignores the C return value and leaves -1
  // on normal return. Require a newly written, parseable stream report too.
  if (code !== 0 && code !== -1) throw new LocalizedError('error.probe');
  const info = JSON.parse(await ff.readFile(jsonPath, 'utf8'));
  if (new URLSearchParams(location.search).has('debug')) console.debug('Probe complete', path);
  if (!info.streams?.length) throw new LocalizedError('error.probe');
  return info;
}
async function convert(job) {
  let ff = await getEngine();
  if (cancelled) throw new LocalizedError('error.stopped');
  if (!available(target(job.options.target))) throw new LocalizedError('error.encoder');
  const input = `input-${job.id}.${inputExtension(job.file.name)}`, output = `output-${job.id}.${target(job.options.target).ext}`, inputInfo = `in-${job.id}.json`, outputInfo = `out-${job.id}.json`;
  logs = [];
  try {
    await ff.writeFile(input, new Uint8Array(await job.file.arrayBuffer()));
    const info = await probe(ff, input, inputInfo); rejectHDR(info, job.options.target);
    // Keep probing and encoding in separate heaps: native entry points retain
    // global state and a failed codec must not poison later phases or jobs.
    ff.terminate(); engine = null; ff = await getEngine();
    if (cancelled) throw new LocalizedError('error.stopped');
    await ff.writeFile(input, new Uint8Array(await job.file.arrayBuffer()));
    job.command = buildArgs(input, output, job.options);
    const result = await ff.exec(job.command, 300000);
    if (result !== 0) throw new LocalizedError(result === 1 && logs.some(s => /timeout/i.test(s)) ? 'error.timeout' : 'error.convert');
    job.progress = 95; progressUI(job);
    const encoded = await ff.readFile(output);
    ff.terminate(); engine = null; ff = await getEngine();
    if (cancelled) throw new LocalizedError('error.stopped');
    await ff.writeFile(output, encoded);
    const resultInfo = await probe(ff, output, outputInfo);
    if (!resultInfo.streams?.length) throw new LocalizedError('error.output');
    const data = await ff.readFile(output);
    if (!(data instanceof Uint8Array) || !data.length) throw new LocalizedError('error.emptyOutput');
    if (cancelled) throw new LocalizedError('error.stopped');
    job.sha256 = [...new Uint8Array(await crypto.subtle.digest('SHA-256', data))].map(v => v.toString(16).padStart(2, '0')).join('');
    if (cancelled) throw new LocalizedError('error.stopped');
    revoke(job); job.output = outputName(job.file.name, job.options.target, job.id); job.outputSize = data.length;
    job.url = URL.createObjectURL(new Blob([data], { type: target(job.options.target).mime }));
    job.verificationKey = 'job.verification';
    job.state = 'completed'; job.progress = 100;
  } catch (error) {
    job.log = logs.join('\n').slice(-8000);
    throw error;
  } finally {
    if (ff.loaded) for (const path of [input, output, inputInfo, outputInfo]) { try { await ff.deleteFile(path); } catch { /* May not exist after a failed conversion. */ } }
  }
}
async function run() {
  if (busy) return;
  busy = true; cancelled = false; notice(); render();
  try {
    for (const job of jobs.filter(j => j.state !== 'completed')) {
      if (cancelled) break;
      current = job; job.state = 'running'; job.error = ''; job.progress = 0; render();
      job.timedOut = false;
      const watchdog = setTimeout(() => { job.timedOut = true; engine?.terminate(); engine = null; }, 300000);
      try { await convert(job); }
      catch (error) {
        job.state = cancelled ? 'cancelled' : 'failed';
        job.error = cancelled ? '' : job.timedOut ? message('error.timeout') : error instanceof LocalizedError ? message(error.key, error.values) : message('error.unknown');
        if (!cancelled && !(error instanceof LocalizedError)) job.log = `${job.log || ''}\n${error?.message || String(error)}`.slice(-8000);
        if (!engine?.loaded && !cancelled) { setEngineState('failed'); notice(message('error.engine')); }
      }
      finally { clearTimeout(watchdog); }
      // Isolate FFmpeg/FFprobe native global state and release the WASM heap
      // between files, including after parser/encoder failures. Assets cache.
      engine?.terminate(); engine = null;
      render();
    }
  } finally { busy = false; current = null; if (!cancelled) setEngineState('finished'); render(); }
}
$('fileInput').addEventListener('change', event => { addFiles(event.target.files); event.target.value = ''; });
$('dropzone').addEventListener('click', () => $('fileInput').click());
for (const event of ['dragenter', 'dragover']) $('dropzone').addEventListener(event, e => { e.preventDefault(); if (!busy) $('dropzone').classList.add('drag'); });
for (const event of ['dragleave', 'drop']) $('dropzone').addEventListener(event, e => { e.preventDefault(); $('dropzone').classList.remove('drag'); if (event === 'drop') addFiles(e.dataTransfer.files); });
window.addEventListener('dragover', e => e.preventDefault());
window.addEventListener('drop', e => e.preventDefault());
$('categories').addEventListener('click', event => { const button = event.target.closest('[data-kind]'); if (!button) return; kind = button.dataset.kind; for (const b of $('categories').children) { b.classList.toggle('active', b === button); b.setAttribute('aria-pressed', String(b === button)); } defaults(); });
$('target').addEventListener('change', hint);
$('apply').addEventListener('click', () => { try { const o = options(); for (const job of jobs.filter(j => j.state !== 'completed')) { job.options = { ...o }; job.state = 'ready'; job.error = ''; } notice(); render(); } catch (error) { notice(error); } });
$('clear').addEventListener('click', () => { jobs.forEach(revoke); jobs.length = 0; notice(); render(); });
$('startButton').addEventListener('click', run);
$('stopButton').addEventListener('click', () => { cancelled = true; engine?.terminate(); engine = null; setEngineState('stopped'); });
$('export').addEventListener('click', () => {
  const results = jobs.map(({ file, url, error, verificationKey, ...job }) => ({ input: file.name, inputBytes: file.size, ...job, error: messageText(error), ...(verificationKey ? { verification: t(verificationKey) } : {}) }));
  const blob = new Blob([JSON.stringify({ schema: 'prism-online/1', version: '2.1.0-beta.1', language: getLanguage(), exportedAt: new Date().toISOString(), results }, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob), link = document.createElement('a'); link.href = url; link.download = 'PrismTranscode-results.json'; link.click(); setTimeout(() => URL.revokeObjectURL(url), 1000);
});
window.addEventListener('beforeunload', event => { if (busy || jobs.some(j => j.url)) { event.preventDefault(); event.returnValue = ''; } });
$('languageSelect').addEventListener('change', event => {
  languagePreference = event.target.value;
  try { if (languagePreference === 'auto') localStorage.removeItem(LANGUAGE_KEY); else localStorage.setItem(LANGUAGE_KEY, languagePreference); } catch { /* Switching still works for this visit. */ }
  updateLanguage();
});
window.addEventListener('languagechange', () => { if (languagePreference === 'auto') updateLanguage(); });
updateLanguage();
if (!browserSupported) notice(message('error.browser'));
