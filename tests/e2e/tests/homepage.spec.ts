import { test, expect } from '@playwright/test';

test.describe('Homepage', () => {
  test('should load the homepage', async ({ page }) => {
    // Navigate to the homepage
    await page.goto('/');

    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Check that we're on the setup page (ShellHub redirects to setup on first load)
    expect(page.url()).toMatch(/\/(setup)?$/);

    // Take a screenshot for debugging
    await page.screenshot({ path: 'homepage.png' });
  });

  test('should have ShellHub title or branding', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Check for ShellHub branding (title or heading)
    const title = await page.title();
    expect(title.toLowerCase()).toMatch(/shellhub|shell hub/i);
  });
});
