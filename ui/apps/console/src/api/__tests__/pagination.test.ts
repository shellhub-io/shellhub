import { describe, it, expect } from "vitest";
import { totalCount } from "../pagination";

describe("totalCount", () => {
  it("returns 0 for undefined", () => {
    expect(totalCount(undefined)).toBe(0);
  });

  it("returns 0 when the property is absent", () => {
    expect(totalCount([1, 2, 3])).toBe(0);
  });

  it("reads a non-enumerable totalCount property", () => {
    const data = [1, 2, 3];
    Object.defineProperty(data, "totalCount", {
      value: 42,
      enumerable: false,
    });
    expect(totalCount(data)).toBe(42);
  });
});
