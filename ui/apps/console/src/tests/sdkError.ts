/**
 * Builds a value shaped like what the SDK throws: the parsed error body with the `status` the
 * fetch error interceptor attaches.
 */
export function sdkError(status: number, body: Record<string, unknown> = {}) {
  // An Error instance, so it also satisfies hooks whose error type is the built-in Error.
  return Object.assign(new Error(`HTTP ${status}`), { ...body, status });
}
