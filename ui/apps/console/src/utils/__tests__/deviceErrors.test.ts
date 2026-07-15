import { describe, it, expect, vi, beforeEach } from "vitest";
import { getConfig, defaultConfig } from "@/env";
import { getAcceptDeviceErrorMessage } from "../deviceErrors";

const mockGetConfig = vi.mocked(getConfig);

describe("getAcceptDeviceErrorMessage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetConfig.mockReturnValue({ ...defaultConfig });
  });

  describe("402 Payment Required", () => {
    it("returns license copy for enterprise (non-cloud) on 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const msg = getAcceptDeviceErrorMessage({ status: 402 });
      expect(msg).toMatch(/license/i);
    });

    it("returns billing copy for cloud on 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const msg = getAcceptDeviceErrorMessage({ status: 402 });
      expect(msg).toMatch(/billing|subscription|plan/i);
    });

    it("cloud 402 message is distinct from enterprise 402 message", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const cloudMsg = getAcceptDeviceErrorMessage({ status: 402 });

      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const enterpriseMsg = getAcceptDeviceErrorMessage({ status: 402 });

      expect(cloudMsg).not.toBe(enterpriseMsg);
    });

    it("returns generic fallback for community on 402 (not billing copy)", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
      });
      const msg = getAcceptDeviceErrorMessage({ status: 402 });
      // Should NOT match cloud billing copy
      expect(msg).not.toMatch(/billing|subscription|plan/i);
    });
  });

  describe("403 Forbidden", () => {
    it("returns namespace copy for 403", () => {
      const msg = getAcceptDeviceErrorMessage({ status: 403 });
      expect(msg).toMatch(/namespace|permission/i);
    });

    it("returns namespace copy even when enterprise is true (403 ignores enterprise flag)", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const msg = getAcceptDeviceErrorMessage({ status: 403 });
      expect(msg).toMatch(/namespace|permission/i);
    });
  });

  describe("409 Conflict", () => {
    it("returns rename copy for 409", () => {
      const msg = getAcceptDeviceErrorMessage({ status: 409 });
      expect(msg).toMatch(/rename|name|already exists/i);
    });
  });

  describe("unknown errors", () => {
    it("returns a generic fallback for an unrecognized status code", () => {
      const msg = getAcceptDeviceErrorMessage({ status: 500 });
      expect(msg).toBeTruthy();
      expect(typeof msg).toBe("string");
    });

    it("returns a generic fallback for a non-sdk error object", () => {
      const msg = getAcceptDeviceErrorMessage(new Error("network failure"));
      expect(msg).toBeTruthy();
      expect(typeof msg).toBe("string");
    });
  });
});
