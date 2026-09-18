import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './online/tests', testMatch: '**/*.spec.js', timeout: 180000, workers: 1,
  use: { baseURL: process.env.TEST_BASE_URL || 'http://127.0.0.1:4173', browserName: 'chromium', acceptDownloads: true, trace: 'retain-on-failure' },
  webServer: process.env.TEST_BASE_URL ? undefined : { command: 'npm run preview', url: 'http://127.0.0.1:4173', reuseExistingServer: !process.env.CI },
  reporter: [['list'], ['html', { open: 'never' }]],
});
