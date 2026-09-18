const { defineConfig } = require('@playwright/test');

if (!process.env.BASE_URL) {
  throw new Error('Set BASE_URL to https://$STAGING_HOST');
}

module.exports = defineConfig({
  testDir: './',
  testMatch: 'e2e.spec.cjs',
  timeout: 30000,
  expect: { timeout: 10000 },
  retries: 1,
  workers: 1,
  reporter: 'list',
  use: {
    headless: true,
    ignoreHTTPSErrors: true,
    baseURL: process.env.BASE_URL,
  },
});
