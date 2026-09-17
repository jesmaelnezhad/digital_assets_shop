const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './',
  timeout: 30000,
  expect: { timeout: 10000 },
  use: { headless: true },
  reporter: 'list',
});
