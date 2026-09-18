import { test, expect } from '@playwright/test';
import { readFile } from 'node:fs/promises';
import { LANGUAGE_KEY } from '../i18n.js';

test.describe('automatic language', () => {
  for (const [locale, expected] of [['en-GB', 'en'], ['zh-TW', 'zh-CN'], ['fr-FR', 'en']]) {
    test.describe(locale, () => {
      test.use({ locale });
      test('follows the browser on first visit', async ({ page }) => {
        await page.goto('/');
        await expect(page.locator('html')).toHaveAttribute('lang', expected);
        await expect(page.locator('#languageSelect')).toHaveValue('auto');
        await expect(page.locator('#startButton')).toContainText(expected === 'en' ? 'Start converting' : '开始转换');
        await expect(page.locator('#guideLink')).toHaveAttribute('href', new RegExp(`USAGE\.${expected === 'en' ? 'en' : 'zh-CN'}\.md$`));
        if (expected === 'en') {
          await expect(page).toHaveTitle('PrismTranscode — Private online media workbench');
          for (const text of await page.locator('[data-i18n]').allTextContents()) expect(text).not.toMatch(/[\u3400-\u9fff]/);
        }
      });
    });
  }
});
test.describe('manual language selection', () => {
  test.use({ locale: 'en-US' });
  test('preference survives reload; browser default removes the override', async ({ page }) => {
    await page.goto('/'); await page.selectOption('#languageSelect', 'zh');
    await expect(page.locator('html')).toHaveAttribute('lang', 'zh-CN');
    await page.reload(); await expect(page.locator('#languageSelect')).toHaveValue('zh');
    await expect(page.locator('#startButton')).toContainText('开始转换');
    await page.selectOption('#languageSelect', 'auto');
    await expect(page.locator('html')).toHaveAttribute('lang', 'en');
    expect(await page.evaluate(key => localStorage.getItem(key), LANGUAGE_KEY)).toBeNull();
    await page.reload(); await expect(page.locator('html')).toHaveAttribute('lang', 'en');
  });
  test('storage denied does not break startup or manual switching', async ({ page }) => {
    const errors = []; page.on('pageerror', error => errors.push(error.message));
    await page.addInitScript(() => { for (const key of ['getItem', 'setItem', 'removeItem']) Storage.prototype[key] = () => { throw new DOMException('Blocked', 'SecurityError'); }; });
    await page.goto('/'); await expect(page.locator('html')).toHaveAttribute('lang', 'en');
    await page.selectOption('#languageSelect', 'zh'); await expect(page.locator('html')).toHaveAttribute('lang', 'zh-CN');
    await page.selectOption('#languageSelect', 'auto'); await expect(page.locator('html')).toHaveAttribute('lang', 'en');
    expect(errors).toEqual([]);
  });
  test('browser language changes affect only automatic mode', async ({ page }) => {
    await page.goto('/');
    await page.evaluate(() => { Object.defineProperty(navigator, 'languages', { configurable: true, value: ['zh-CN'] }); window.dispatchEvent(new Event('languagechange')); });
    await expect(page.locator('html')).toHaveAttribute('lang', 'zh-CN');
    await page.selectOption('#languageSelect', 'en');
    await page.evaluate(() => window.dispatchEvent(new Event('languagechange')));
    await expect(page.locator('html')).toHaveAttribute('lang', 'en');
    await page.selectOption('#languageSelect', 'auto');
    await expect(page.locator('html')).toHaveAttribute('lang', 'zh-CN');
  });
  test('license links serve the selected language', async ({ page, request }) => {
    await page.goto('/');
    for (const language of ['en', 'zh']) {
      await page.selectOption('#languageSelect', language);
      const response = await request.get(await page.locator('#licenseLink').getAttribute('href'));
      expect(response.ok()).toBeTruthy();
      expect(await response.text()).toContain(language === 'en' ? '# Third-party sources and licenses' : '# 技术来源与第三方说明');
    }
  });
  test('live queue, settings, conversion, downloads and errors survive switching', async ({ page }) => {
    const errors = []; page.on('pageerror', error => errors.push(error.message));
    await page.goto('/'); await page.selectOption('#target', 'wav'); await page.selectOption('#quality', 'high');
    await page.locator('#fileInput').setInputFiles('testdata/tone.mp3');
    await page.selectOption('#languageSelect', 'zh');
    await expect(page.locator('.job-select')).toHaveValue('wav'); await expect(page.locator('#quality')).toHaveValue('high');
    await expect(page.locator('.job-status')).toHaveText('等待转换');
    await page.locator('#startButton').click(); await expect(page.locator('.job')).toHaveAttribute('data-state', 'running');
    await page.selectOption('#languageSelect', 'en');
    await expect(page.locator('.job')).toHaveAttribute('data-state', 'completed', { timeout: 120000 });
    await expect(page.locator('.job-status')).toHaveText('Completed');
    const originalURL = await page.locator('.job-actions a').getAttribute('href');
    await page.selectOption('#languageSelect', 'zh');
    expect(await page.locator('.job-actions a').getAttribute('href')).toBe(originalURL);
    await page.selectOption('#languageSelect', 'en');
    const downloadEvent = page.waitForEvent('download'); await page.getByRole('link', { name: 'Download tone_prism_1.wav' }).click();
    expect((await readFile(await (await downloadEvent).path())).subarray(0, 4).toString()).toBe('RIFF');
    const reportEvent = page.waitForEvent('download'); await page.locator('#export').click();
    const report = JSON.parse(await readFile(await (await reportEvent).path(), 'utf8'));
    expect(report.language).toBe('en'); expect(report.results[0].verification).toContain('output structure check');
    await page.locator('#fileInput').setInputFiles({ name: 'sample.ncm', mimeType: 'application/octet-stream', buffer: Buffer.from('test') });
    await expect(page.getByRole('alert')).toContainText('Use the desktop app for NCM');
    await page.selectOption('#languageSelect', 'zh'); await expect(page.getByRole('alert')).toContainText('NCM 请使用桌面版');
    await page.locator('#fileInput').setInputFiles({ name: 'broken.mp3', mimeType: 'audio/mpeg', buffer: Buffer.from('not media') });
    await page.locator('#startButton').click(); await expect(page.locator('.job').last()).toHaveAttribute('data-state', 'failed', { timeout: 60000 });
    await page.selectOption('#languageSelect', 'en');
    expect(await page.locator('.job-status').last().textContent()).not.toMatch(/[\u3400-\u9fff]/);
    expect(errors).toEqual([]);
  });
  test('language control and layouts remain usable in both languages', async ({ page }) => {
    await page.goto('/');
    for (const language of ['en', 'zh']) {
      await page.selectOption('#languageSelect', language);
      for (const width of [1440, 768, 390, 320]) {
        await page.setViewportSize({ width, height: 900 });
        await expect(page.locator('#languageSelect')).toBeVisible();
        expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `${language} at ${width}px`).toBeTruthy();
        if (width === 1440 || width === 390) await page.screenshot({ path: `artifacts/online-${language}-${width}.png`, fullPage: true });
      }
    }
  });
});
