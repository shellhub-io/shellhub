import axios from "axios";
import { expect } from "vitest";

declare module "vitest" {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  interface Assertion<T = any> {
    toBeAxiosErrorWithStatus(expectedStatus: number): T;
  }
}

expect.extend({
  toBeAxiosErrorWithStatus(received: unknown, expectedStatus: number) {
    const isAxiosError = axios.isAxiosError(received);

    if (!isAxiosError) {
      return {
        pass: false,
        message: () =>
          `Expected an AxiosError with response, but received: ${typeof received}\n`
          + `${received instanceof Error ? `Error message: ${received.message}` : `Value: ${String(received)}`}`,
      };
    }

    const actualStatus = received.response?.status;
    const pass = Number(actualStatus) === expectedStatus;

    return {
      pass,
      message: () =>
        pass
          ? `Expected error status NOT to be ${expectedStatus}`
          : `Expected error status ${expectedStatus}, but received ${actualStatus}`,
    };
  },
});
