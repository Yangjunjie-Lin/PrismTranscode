import { test, expect } from '@playwright/test';
import { readFile } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import path from 'node:path';
import { TARGETS } from '../convert.js';
test.use({ locale: 'zh-CN' });
const fixture = path.resolve('testdata/tone.mp3');
test('all 14 output presets convert real synthetic media', async ({ page }) => {
  if (process.env.PRISM_DEBUG) page.on('console', message => { if (!/^\s*[VAS. ][FSXBD. ]{5}/.test(message.text())) console.log(message.text()); });
  await page.goto(process.env.PRISM_DEBUG ? '/?debug' : '/');
  const files = [];
  for (const t of TARGETS) {
    const source = t.kind === 'audio' ? 'tone.mp3' : t.kind === 'video' || t.id === 'gif' ? 'clip.mp4' : 'image.png';
    files.push({ name: `${t.id}-${t.kind === 'subtitle' ? 'captions.srt' : source}`, mimeType: 'application/octet-stream', buffer: t.kind === 'subtitle' ? Buffer.from('1\n00:00:00,000 --> 00:00:01,000\nPrism\n') : await readFile(path.resolve('testdata', source)) });
  }
  await page.locator('#fileInput').setInputFiles(files);
  for (let i = 0; i < TARGETS.length; i++) await page.locator('.job-select').nth(i).selectOption(TARGETS[i].id);
  await page.locator('#startButton').click();
  await expect(page.locator('#startButton')).toBeVisible({ timeout: 150000 });
  const statuses = await page.locator('.job-status').allTextContents();
  if (await page.locator('.job[data-state="failed"]').count()) {
    const reportEvent = page.waitForEvent('download'); await page.locator('#export').click();
    const report = JSON.parse(await readFile(await (await reportEvent).path(), 'utf8'));
    console.log(report.results.filter(j => j.state === 'failed').map(j => ({ input: j.input, error: j.error, log: j.log })));
  }
  expect(await page.locator('.job[data-state="completed"]').count(), statuses.join('\n')).toBe(14);
  await expect(page.getByRole('link', { name: /^下载 .*_prism_/ })).toHaveCount(14);
});
test('real MP3 → WAV conversion, verified download, report and no media upload', async ({ page }) => {
  const errors = [], writes = [];
  page.on('pageerror', e => errors.push(e.message));
  page.on('request', r => { if (['POST', 'PUT', 'PATCH'].includes(r.method())) writes.push(r.url()); });
  await page.goto('/');
  await page.selectOption('#target', 'wav');
  await page.locator('#fileInput').setInputFiles(fixture);
  await page.getByRole('button', { name: '开始转换' }).click();
  await expect(page.locator('.job')).toHaveAttribute('data-state', /completed|failed/, { timeout: 120000 });
  await expect(page.locator('.job'), await page.locator('.job-status').innerText()).toHaveAttribute('data-state', 'completed');
  const downloadEvent = page.waitForEvent('download');
  await page.getByRole('link', { name: /下载 tone_prism/ }).click();
  const download = await downloadEvent, data = await readFile(await download.path());
  expect(data.subarray(0, 4).toString()).toBe('RIFF'); expect(data.subarray(8, 12).toString()).toBe('WAVE'); expect(data.length).toBeGreaterThan(1000);
  const reportEvent = page.waitForEvent('download'); await page.locator('#export').click();
  const report = JSON.parse(await readFile(await (await reportEvent).path(), 'utf8'));
  expect(report.results[0].sha256).toBe(createHash('sha256').update(data).digest('hex'));
  expect(writes).toEqual([]); expect(errors).toEqual([]);
  await page.screenshot({ path: 'artifacts/online-desktop.png', fullPage: true });
});
test('failure isolation, per-file targets, subtitle conversion and retry', async ({ page }) => {
  await page.goto('/'); await page.selectOption('#target', 'flac');
  await page.locator('#fileInput').setInputFiles([{ name: 'broken.mp3', mimeType: 'audio/mpeg', buffer: Buffer.from('not media') }, { name: 'good.mp3', mimeType: 'audio/mpeg', buffer: await readFile(fixture) }, { name: 'captions.srt', mimeType: 'text/plain', buffer: Buffer.from('1\n00:00:00,000 --> 00:00:01,000\nHello Prism\n') }]);
  await page.getByLabel('captions.srt 的输出格式').selectOption('vtt');
  await page.locator('#startButton').click();
  await expect(page.locator('.job').nth(0)).toHaveAttribute('data-state', 'failed', { timeout: 120000 });
  await expect(page.locator('.job').nth(1)).toHaveAttribute('data-state', 'completed', { timeout: 120000 });
  await expect(page.locator('.job').nth(2)).toHaveAttribute('data-state', 'completed', { timeout: 120000 });
  const ev = page.waitForEvent('download'); await page.getByRole('link', { name: /下载 captions/ }).click();
  expect(await readFile(await (await ev).path(), 'utf8')).toContain('WEBVTT');
  await page.locator('#startButton').click(); await expect(page.locator('#startButton')).toBeVisible();
  await expect(page.locator('.job[data-state="completed"]')).toHaveCount(2);
});
test('stop during initialization and then retry succeeds', async ({ page }) => {
  await page.goto('/'); await page.selectOption('#target', 'wav'); await page.locator('#fileInput').setInputFiles(fixture);
  await page.locator('#startButton').click(); await page.locator('#stopButton').click();
  await expect(page.locator('.job')).toHaveAttribute('data-state', 'cancelled');
  await page.locator('#startButton').click(); await expect(page.locator('.job')).toHaveAttribute('data-state', 'completed', { timeout: 120000 });
});
test('NCM rejection, deduplication, hostile names and mobile layout', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 }); await page.goto('/');
  await page.locator('#fileInput').setInputFiles({ name: 'sample.ncm', mimeType: 'application/octet-stream', buffer: Buffer.from('ncm') });
  await expect(page.getByRole('alert')).toContainText('NCM 请使用桌面版');
  await page.locator('#fileInput').setInputFiles(fixture); await page.locator('#fileInput').setInputFiles(fixture); await expect(page.locator('.job')).toHaveCount(1);
  await page.locator('#fileInput').setInputFiles({ name: '<img onerror=alert(1)>.mp3', mimeType: 'audio/mpeg', buffer: Buffer.from('test') });
  await expect(page.locator('.job-name').last()).toHaveText('<img onerror=alert(1)>.mp3'); expect(await page.locator('#queue img').count()).toBe(0);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBeTruthy();
  await page.screenshot({ path: 'artifacts/online-mobile.png', fullPage: true });
  await page.locator('#clear').click(); await expect(page.locator('.job')).toHaveCount(0);
});
