import { describe, it, expect, beforeEach } from "vitest";
import {
  getSafeRedirect,
  setPendingDeviceCode,
  clearPendingDeviceCode,
  consumePendingDeviceCode,
  hasPendingDeviceCode,
  resolvePostLoginRedirect,
  PENDING_DEVICE_CODE_KEY,
} from "../navigation";

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
      expect(getSafeRedirect(params("/settings/profile"))).toBe(
        "/settings/profile",
      );
    });

    it("returns a path with a query string", () => {
      expect(getSafeRedirect(params("/devices?page=2"))).toBe(
        "/devices?page=2",
      );
    });

    it("returns a path with a hash fragment", () => {
      expect(getSafeRedirect(params("/namespaces#list"))).toBe(
        "/namespaces#list",
      );
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

describe("pending device code", () => {
  beforeEach(() => {
    localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
  });

  it("setPendingDeviceCode stores the code", () => {
    setPendingDeviceCode("WXYZ2K7Q");
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBe("WXYZ2K7Q");
  });

  it("clearPendingDeviceCode removes the key", () => {
    setPendingDeviceCode("ABC");
    clearPendingDeviceCode();
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
  });

  it("consumePendingDeviceCode returns and removes the code", () => {
    setPendingDeviceCode("WXYZ2K7Q");
    expect(consumePendingDeviceCode()).toBe("WXYZ2K7Q");
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
  });

  it("consumePendingDeviceCode returns null when no key exists", () => {
    expect(consumePendingDeviceCode()).toBeNull();
  });

  it("a second setPendingDeviceCode overwrites the first", () => {
    setPendingDeviceCode("OLD");
    setPendingDeviceCode("NEW");
    expect(consumePendingDeviceCode()).toBe("NEW");
  });

  it("hasPendingDeviceCode returns true when a code exists", () => {
    setPendingDeviceCode("WXYZ2K7Q");
    expect(hasPendingDeviceCode()).toBe(true);
  });

  it("hasPendingDeviceCode returns false when no code exists", () => {
    expect(hasPendingDeviceCode()).toBe(false);
  });
});

describe("resolvePostLoginRedirect", () => {
  beforeEach(() => {
    localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
  });

  it("returns /accept-device with the pending code when no explicit redirect", () => {
    setPendingDeviceCode("WXYZ2K7Q");
    expect(resolvePostLoginRedirect(new URLSearchParams())).toBe(
      "/accept-device?code=WXYZ2K7Q",
    );
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
  });

  it("returns the explicit redirect and leaves the code untouched", () => {
    setPendingDeviceCode("WXYZ2K7Q");
    expect(resolvePostLoginRedirect(params("/devices"))).toBe("/devices");
    expect(hasPendingDeviceCode()).toBe(true);
  });

  it("returns /dashboard when no redirect and no pending code", () => {
    expect(resolvePostLoginRedirect(new URLSearchParams())).toBe("/dashboard");
  });

  it("encodes special characters in the code", () => {
    setPendingDeviceCode("A&B=C");
    expect(resolvePostLoginRedirect(new URLSearchParams())).toBe(
      "/accept-device?code=A%26B%3DC",
    );
  });
});
