import { describe, it, expect } from "vitest";
import { queryClient } from "../queryClient";

function shouldRetry(failureCount: number, error: Error): boolean {
  const retry = queryClient.getDefaultOptions().queries?.retry;
  if (typeof retry !== "function") {
    throw new Error(`expected a retry function, got ${String(retry)}`);
  }
  return retry(failureCount, error);
}

const httpError = (status: number) =>
  Object.assign(new Error(), { status, headers: new Headers() });

describe("queryClient retry", () => {
  it.each([400, 403, 404])("does not retry a %i", (status) => {
    expect(shouldRetry(0, httpError(status))).toBe(false);
  });

  it("retries a server error up to three times", () => {
    expect(shouldRetry(2, httpError(500))).toBe(true);
    expect(shouldRetry(3, httpError(500))).toBe(false);
  });

  it("retries a request that got no response", () => {
    expect(shouldRetry(0, new TypeError("Failed to fetch"))).toBe(true);
  });
});
