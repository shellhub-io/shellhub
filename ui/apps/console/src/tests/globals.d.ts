import type { Mock } from "vitest";

declare global {
  function mockSdkGen<T extends Record<string, Mock>>(mocks: T): T;
}
