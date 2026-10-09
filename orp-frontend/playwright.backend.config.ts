import type { PlaywrightTestConfig } from '@playwright/test';

const config: PlaywrightTestConfig = {
  expect: { timeout: 8000 },
  reporter: [['list'], ['html', { outputFolder: 'output/backend-ui-playwright' }]],
  testDir: './playground/__tests__/e2e/backend-ui',
  timeout: 15 * 60 * 1000,
  use: {
    baseURL: 'http://127.0.0.1:5666',
    headless: false,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  workers: 1,
};

export default config;
