import { describe, it, expect } from "vitest";
import { getSafeRedirect } from "../navigation";

function params(value?: string): URLSearchParams {
  const p = new URLSearchParams();
  if (value !== undefined) p.set("redirect", value);
  return p;
}

describe("getSafeRedirect", () => {
  describe("safe paths — returned as-is", () => {
    it("returns a simple absolute path", () => {
      expect(getSafeRedirect(params("/dashboard"))).toBe("/dashboard");
    });

    it("returns a nested path", () => {
      expect(getSafeRedirect(params("/settings/profile"))).toBe("/settings/profile");
    });

    it("returns a path with a query string", () => {
      expect(getSafeRedirect(params("/devices?page=2"))).toBe("/devices?page=2");
    });

    it("returns a path with a hash fragment", () => {
      expect(getSafeRedirect(params("/namespaces#list"))).toBe("/namespaces#list");
    });
  });

  describe("unsafe paths — fall back to default", () => {
    it("rejects protocol-relative URL starting with //", () => {
      expect(getSafeRedirect(params("//evil.com"))).toBe("/dashboard");
    });

    it("rejects backslash trick starting with /\\", () => {
      expect(getSafeRedirect(params("/\\evil.com"))).toBe("/dashboard");
    });

    it("rejects a bare domain without leading slash", () => {
      expect(getSafeRedirect(params("evil.com/path"))).toBe("/dashboard");
    });

    it("rejects an http:// URL", () => {
      expect(getSafeRedirect(params("http://evil.com"))).toBe("/dashboard");
    });

    it("rejects an https:// URL", () => {
      expect(getSafeRedirect(params("https://evil.com"))).toBe("/dashboard");
    });
  });

  describe("missing or empty param", () => {
    it("falls back to /dashboard when redirect param is absent", () => {
      expect(getSafeRedirect(new URLSearchParams())).toBe("/dashboard");
    });

    it("falls back to /dashboard when redirect param is empty string", () => {
      expect(getSafeRedirect(params(""))).toBe("/dashboard");
    });
  });

  describe("custom fallback", () => {
    it("uses the provided fallback when redirect param is absent", () => {
      expect(getSafeRedirect(new URLSearchParams(), "/home")).toBe("/home");
    });

    it("uses the provided fallback when redirect is unsafe", () => {
      expect(getSafeRedirect(params("//evil.com"), "/home")).toBe("/home");
    });

    it("ignores the custom fallback when redirect is safe", () => {
      expect(getSafeRedirect(params("/devices"), "/home")).toBe("/devices");
    });
  });
});
