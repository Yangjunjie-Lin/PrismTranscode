import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import { messages, resolveLanguage, setLanguage, t, messageText, LocalizedError } from '../i18n.js';
import { TARGETS, validateOptions, fileError, rejectHDR } from '../convert.js';

test('browser language selection, regional variants, priority and fallback', () => {
  for (const lang of ['zh', 'zh-CN', 'zh-TW', 'zh-HK', 'zh-Hant', 'ZH_cn']) assert.equal(resolveLanguage('auto', [lang]), 'zh');
  for (const lang of ['en', 'en-US', 'en-GB']) assert.equal(resolveLanguage('auto', [lang]), 'en');
  assert.equal(resolveLanguage('auto', ['en-US', 'zh-CN']), 'en');
  assert.equal(resolveLanguage('auto', ['fr-FR', 'zh-CN', 'en']), 'zh');
  assert.equal(resolveLanguage('auto', ['fr', 'de']), 'en');
  assert.equal(resolveLanguage('invalid', []), 'en');
  assert.equal(resolveLanguage('zh', ['en']), 'zh');
  assert.equal(resolveLanguage('en', ['zh']), 'en');
});
test('all messages have two translations and matching placeholders', () => {
  for (const [key, pair] of Object.entries(messages)) {
    assert.equal(pair.length, 2, key);
    assert.ok(pair.every(value => typeof value === 'string' && value.trim()), key);
    assert.deepEqual(pair[0].match(/\{\w+\}/g)?.sort() || [], pair[1].match(/\{\w+\}/g)?.sort() || [], key);
  }
  for (const format of TARGETS) for (const part of ['label', 'hint']) assert.ok(messages[`format.${format.id}.${part}`]);
  const html = readFileSync(new URL('../index.html', import.meta.url), 'utf8');
  for (const match of html.matchAll(/data-i18n(?:-aria)?="([^"]+)"/g)) assert.ok(messages[match[1]], match[1]);
});
test('errors created in one language retranslate without changing validation', () => {
  setLanguage('zh');
  const err = new LocalizedError('error.time');
  assert.match(messageText(err), /86400 秒/);
  const ncm = fileError({ name: 'test.NCM', size: 30 });
  setLanguage('en');
  assert.match(messageText(err), /86400 seconds/);
  assert.match(messageText(ncm), /desktop app for NCM/);
  assert.throws(() => validateOptions({ target: 'invalid' }), LocalizedError);
  assert.throws(() => rejectHDR({ streams: [{ color_transfer: 'smpte2084' }] }, 'mp4'), LocalizedError);
});
test('interpolated filenames remain literal strings, not translated or parsed', () => {
  assert.equal(t('job.removeLabel', { name: '<img>{count}.mp3' }, 'en'), 'Remove <img>{count}.mp3');
});
test('bilingual documentation links resolve to repository files', () => {
  const root = new URL('../../', import.meta.url);
  for (const name of ['README.md', 'README.en.md', 'DESKTOP.md', 'DESKTOP.en.md', 'SECURITY.md', 'SECURITY.en.md', 'THIRD_PARTY_NOTICES.md', 'THIRD_PARTY_NOTICES.en.md', 'docs/USAGE.zh-CN.md', 'docs/USAGE.en.md']) {
    const url = new URL(name, root), source = readFileSync(url, 'utf8');
    for (const match of source.matchAll(/\]\((?!https?:\/\/|#)([^)#]+)(?:#[^)]*)?\)/g)) assert.ok(existsSync(new URL(match[1], url)), `${name}: ${match[1]}`);
  }
});
