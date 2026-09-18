const { defineConfig } = require('@playwright/test');

if (!process.env.BASE_URL) {
  throw new Error('Set BASE_URL to https://$STAGING_HOST');
}

module.exports = defineConfig({
  testDir: '.',
  testIgnore: ['**/e2e-root.spec.cjs'],
  timeout: 45000,
  retries: 1,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL: process.env.BASE_URL,
    headless: true,
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { browserName: 'chromium' },
    },
  ],
});
