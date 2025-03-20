/**
 * Vitest setup function
 */
export async function setup() {
  global.CSS = {
    supports: () => false,
    escape: (str: string) => str,
  } as never;
}

// FAIL LOUDLY on unhandled promise rejections / errors
process.on("unhandledRejection", (reason) => {
  // eslint-disable-next-line no-console
  console.log("FAILED TO HANDLE PROMISE REJECTION");
  throw reason;
});

/**
 * Vitest Teardown function
 */
export async function teardown() {
  console.log("📝 vitest globalTeardown");
}
