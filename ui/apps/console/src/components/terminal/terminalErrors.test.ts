import { describe, it, expect, vi, beforeEach } from "vitest";
import { resolveError } from "./terminalErrors";
import { getConfig } from "@/env";

const mockGetConfig = vi.mocked(getConfig);

beforeEach(() => {
  mockGetConfig.mockReturnValue({
    edition: "community",
  } as ReturnType<typeof getConfig>);
});

describe("resolveError", () => {
  describe("access to the device has been denied", () => {
    it("resolves with reconnect false", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
      );
      expect(result.reconnect).toBe(false);
    });

    it("uses an 'Access denied' title, not 'Connection failed'", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
      );
      expect(result.title).toBe("Access denied");
    });

    it("includes the permission-denied message", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
      );
      expect(result.message).toContain("permission");
    });

    it("includes a hint about billing or namespace policy on community edition", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
      );
      const hintText = result.hints.join(" ").toLowerCase();
      expect(hintText.includes("billing") || hintText.includes("policy")).toBe(
        true,
      );
    });

    it("does not show the word 'firewall' twice when cloud edition appends its hint", () => {
      mockGetConfig.mockReturnValue({
        edition: "cloud",
      } as ReturnType<typeof getConfig>);
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
      );
      const allHintText = result.hints.join(" ").toLowerCase();
      const count = (allHintText.match(/firewall/g) ?? []).length;
      expect(count).toBeLessThanOrEqual(1);
    });

    it("includes a firewall/rules link when cloud or enterprise is enabled", () => {
      mockGetConfig.mockReturnValue({
        edition: "cloud",
      } as ReturnType<typeof getConfig>);
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
      );
      expect(result.links.some((l) => l.to === "/firewall-rules")).toBe(true);
    });

    it("does not include a firewall/rules link on community edition", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
      );
      expect(result.links.some((l) => l.to === "/firewall-rules")).toBe(false);
    });
  });

  describe("invalid sshid format", () => {
    it("resolves with reconnect false", () => {
      const result = resolveError("invalid sshid format", "uid-2");
      expect(result.reconnect).toBe(false);
    });

    it("uses an 'Invalid connection identifier' title, not 'Connection failed'", () => {
      const result = resolveError("invalid sshid format", "uid-2");
      expect(result.title).toBe("Invalid connection identifier");
    });

    it("includes a hint showing the username@namespace.device@host form", () => {
      const result = resolveError("invalid sshid format", "uid-2");
      const hintText = result.hints.join(" ");
      // Must show the SSHID format pattern
      expect(hintText).toMatch(/@.*\./);
      expect(hintText).toMatch(/@.*@/);
    });
  });

  describe("unknown error key", () => {
    it("returns the generic 'Connection failed' title", () => {
      const result = resolveError("some unknown error", "uid-3");
      expect(result.title).toBe("Connection failed");
    });
  });

  describe("firewall link behavior for other known errors", () => {
    it("does not include firewall/rules link for authentication errors on community edition", () => {
      const result = resolveError("failed to authenticate to device", "uid-4");
      expect(result.links.some((l) => l.to === "/firewall-rules")).toBe(false);
    });

    it("includes firewall/rules link for authentication errors on cloud edition", () => {
      mockGetConfig.mockReturnValue({
        edition: "cloud",
      } as ReturnType<typeof getConfig>);
      const result = resolveError("failed to authenticate to device", "uid-4");
      expect(result.links.some((l) => l.to === "/firewall-rules")).toBe(true);
    });
  });
});
