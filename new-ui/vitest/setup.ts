/**
 * Vitest setup function
 */
export async function setup() {
  global.CSS = {
    supports: (str: string) => false,
    escape: (str: string) => str,
  };
}

/**
 * Vitest Teardown function
 */
export async function teardown() {
  console.log("📝 vitest globalTeardown");
}
