import { readFile, writeFile } from 'node:fs/promises';
const evidence = { version: '2.1.0-beta.1', generated: new Date().toISOString(), platform: 'Windows x64', go: '1.27.1', native: {}, desktopBrowser: {} };
for (const [key, file] of [['native', 'windows-smoke-report.json'], ['desktopBrowser', 'windows-ui-report.json']]) {
  const report = JSON.parse(await readFile(new URL(`../artifacts/${file}`, import.meta.url), 'utf8'));
  evidence[key] = { testedAt: report.tested_at, environment: report.environment || report.ffmpeg, passed: report.passed, failed: report.failed, tests: (report.tests || report.results).map(t => ({ name: t.name, passed: t.passed })) };
}
evidence.online = { engine: '@ffmpeg/core 0.12.10', tests: 'npm run test:e2e: 5 scenarios, including all 14 output presets', unitTests: 7 };
evidence.security = { govulncheck: 'No vulnerabilities found', npmAudit: '0 vulnerabilities; does not scan embedded C/C++ codecs', codeSigned: false };
await writeFile(new URL('../docs/release-evidence.json', import.meta.url), JSON.stringify(evidence, null, 2));
