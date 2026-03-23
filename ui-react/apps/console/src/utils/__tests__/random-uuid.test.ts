import { describe, it, expect, vi, afterEach } from "vitest";
import { generateRandomUUID } from "@/utils/random-uuid";

const UUID_V4_REGEX
  = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

describe("randomUUID", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("returns a valid UUID v4", () => {
    expect(generateRandomUUID()).toMatch(UUID_V4_REGEX);
  });

  it("returns unique values across calls", () => {
    const ids = new Set(Array.from({ length: 50 }, () => generateRandomUUID()));
    expect(ids.size).toBe(50);
  });

  describe("when crypto.randomUUID is unavailable", () => {
    it("falls back to a valid UUID v4", () => {
      vi.stubGlobal("crypto", {
        getRandomValues: crypto.getRandomValues.bind(crypto),
      });

      const uuid = generateRandomUUID();
      expect(uuid).toMatch(UUID_V4_REGEX);
    });

    it("returns unique values across calls", () => {
      vi.stubGlobal("crypto", {
        getRandomValues: crypto.getRandomValues.bind(crypto),
      });

      const ids = new Set(Array.from({ length: 50 }, () => generateRandomUUID()));
      expect(ids.size).toBe(50);
    });
  });
});
