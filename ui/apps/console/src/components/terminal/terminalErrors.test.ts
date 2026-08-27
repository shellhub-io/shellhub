import { describe, it, expect, vi, beforeEach } from "vitest";
import { resolveError } from "./terminalErrors";
import { defaultConfig, getConfig } from "@/env";

const mockGetConfig = vi.mocked(getConfig);

describe("resolveError", () => {
  beforeEach(() => {
    mockGetConfig.mockReturnValue({
      ...defaultConfig,
      edition: "community",
    });
  });
  describe("access to the device has been denied", () => {
    it("resolves with reconnect false", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
        false,
      );
      expect(result.reconnect).toBe(false);
    });

    it("uses an 'Access denied' title, not 'Connection failed'", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
        false,
      );
      expect(result.title).toBe("Access denied");
    });

    it("includes the permission-denied message", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
        false,
      );
      expect(result.message).toContain("permission");
    });

    it("includes a hint about billing or namespace policy on community edition", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
        false,
      );
      const hintText = result.hints.join(" ").toLowerCase();
      expect(hintText.includes("billing") || hintText.includes("policy")).toBe(
        true,
      );
    });

    it("does not show the word 'firewall' twice when cloud edition appends its hint", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
        false,
      );
      const allHintText = result.hints.join(" ").toLowerCase();
      const count = (allHintText.match(/firewall/g) ?? []).length;
      expect(count).toBeLessThanOrEqual(1);
    });

    it("includes a firewall/rules link when cloud or enterprise is enabled", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
        false,
      );
      expect(result.links.some((l) => l.to === "/firewall-rules")).toBe(true);
    });

    it("does not include a firewall/rules link on community edition", () => {
      const result = resolveError(
        "access to the device has been denied",
        "uid-1",
        false,
      );
      expect(result.links.some((l) => l.to === "/firewall-rules")).toBe(false);
    });

    describe("identity mode", () => {
      beforeEach(() => {
        mockGetConfig.mockReturnValue({
          ...defaultConfig,
          edition: "cloud",
        });
      });

      it("says 'access policy' instead of 'firewall rule' in hints", () => {
        const result = resolveError(
          "access to the device has been denied",
          "uid-1",
          true,
        );
        const hintText = result.hints.join(" ").toLowerCase();
        expect(hintText).toContain("access policy");
        expect(hintText).not.toContain("firewall");
      });

      it("labels the link 'Access policies' instead of 'Firewall rules'", () => {
        const result = resolveError(
          "access to the device has been denied",
          "uid-1",
          true,
        );
        const link = result.links.find((l) => l.to === "/firewall-rules");
        expect(link).toBeDefined();
        expect(link!.label).toBe("Access policies");
      });

      it("says 'access policy' instead of 'namespace policy' in the base hint", () => {
        const result = resolveError(
          "access to the device has been denied",
          "uid-1",
          true,
        );
        const hintText = result.hints.join(" ").toLowerCase();
        expect(hintText).toContain("access policy");
        expect(hintText).not.toContain("namespace policy");
      });

      it("keeps firewall wording when not in identity mode", () => {
        const result = resolveError(
          "access to the device has been denied",
          "uid-1",
          false,
        );
        const hintText = result.hints.join(" ").toLowerCase();
        expect(hintText).toContain("firewall");
        expect(result.links.some((l) => l.label === "Firewall rules")).toBe(
          true,
        );
      });
    });
  });

  describe("invalid sshid format", () => {
    it("resolves with reconnect false", () => {
      const result = resolveError("invalid sshid format", "uid-2", false);
      expect(result.reconnect).toBe(false);
    });

    it("uses an 'Invalid connection identifier' title, not 'Connection failed'", () => {
      const result = resolveError("invalid sshid format", "uid-2", false);
      expect(result.title).toBe("Invalid connection identifier");
    });

    it("includes a hint showing the username@namespace.device@host form", () => {
      const result = resolveError("invalid sshid format", "uid-2", false);
      const hintText = result.hints.join(" ");
      // Must show the SSHID format pattern
      expect(hintText).toMatch(/@.*\./);
      expect(hintText).toMatch(/@.*@/);
    });
  });

  describe("unknown error key", () => {
    it("returns the generic 'Connection failed' title", () => {
      const result = resolveError("some unknown error", "uid-3", false);
      expect(result.title).toBe("Connection failed");
    });
  });

  describe("firewall link behavior for other known errors", () => {
    it("does not include firewall/rules link for authentication errors on community edition", () => {
      const result = resolveError(
        "failed to authenticate to device",
        "uid-4",
        false,
      );
      expect(result.links.some((l) => l.to === "/firewall-rules")).toBe(false);
    });

    it("includes firewall/rules link for authentication errors on cloud edition", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });

      const result = resolveError(
        "failed to authenticate to device",
        "uid-4",
        false,
      );
      expect(result.links.some((l) => l.to === "/firewall-rules")).toBe(true);
    });

    it("labels the link 'Access policies' in identity mode", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });

      const result = resolveError(
        "failed to authenticate to device",
        "uid-4",
        true,
      );
      const link = result.links.find((l) => l.to === "/firewall-rules");
      expect(link).toBeDefined();
      expect(link!.label).toBe("Access policies");
    });
  });
});
