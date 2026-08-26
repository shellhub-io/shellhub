import { describe, it, expect, vi, beforeEach } from "vitest";
import { getConfig, defaultConfig } from "@/env";
import { getAcceptErrorMessage } from "../acceptErrors";

const mockGetConfig = vi.mocked(getConfig);

const msg = (
  err: unknown,
  hasSubscription = false,
  canSubscribe = true,
  entityType?: "device" | "container",
) => getAcceptErrorMessage(err, hasSubscription, canSubscribe, entityType);

describe("getAcceptErrorMessage", () => {
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
      expect(msg({ status: 402 })).toMatch(/license/i);
    });

    it("returns billing copy for cloud on 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      expect(msg({ status: 402 })).toMatch(/billing|subscription|plan/i);
    });

    it("cloud 402 message is distinct from enterprise 402 message", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const cloudMsg = msg({ status: 402 });

      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const enterpriseMsg = msg({ status: 402 });

      expect(cloudMsg).not.toBe(enterpriseMsg);
    });

    it("separates a subscription block from a device-limit hit on cloud 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });

      const limitMsg = msg({ status: 402 });
      const subscriptionMsg = msg({ status: 402 }, true);

      expect(subscriptionMsg).not.toBe(limitMsg);
      expect(limitMsg).toMatch(/device limit/i);
      expect(subscriptionMsg).not.toMatch(/device limit/i);
    });

    it("returns generic fallback for community on 402 (not billing copy)", () => {
      expect(msg({ status: 402 })).not.toMatch(/billing|subscription|plan/i);
    });
  });

  describe("403 Forbidden", () => {
    it("returns namespace copy for 403", () => {
      expect(msg({ status: 403 })).toMatch(/namespace|permission/i);
    });

    it("returns namespace copy even when enterprise is true (403 ignores enterprise flag)", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      expect(msg({ status: 403 })).toMatch(/namespace|permission/i);
    });
  });

  describe("409 Conflict", () => {
    it("returns rename copy for 409", () => {
      expect(msg({ status: 409 })).toMatch(/rename|name|already exists/i);
    });
  });

  describe("unknown errors", () => {
    it("returns a generic fallback for an unrecognized status code", () => {
      expect(msg({ status: 500 })).toBeTruthy();
    });

    it("returns a generic fallback for a non-sdk error object", () => {
      expect(msg(new Error("network failure"))).toBeTruthy();
    });
  });

  describe("entityType parameter", () => {
    it("interpolates 'container' into enterprise 402 copy", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const result = msg({ status: 402 }, false, true, "container");
      expect(result).toMatch(/container limit/i);
      expect(result).not.toMatch(/device/i);
    });

    it("interpolates 'container' into cloud 402 billing copy", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const result = msg({ status: 402 }, false, true, "container");
      expect(result).toMatch(/container limit/i);
      expect(result).not.toMatch(/device/i);
    });

    it("interpolates 'container' into 403 copy", () => {
      const result = msg({ status: 403 }, false, true, "container");
      expect(result).toMatch(/containers/i);
      expect(result).not.toMatch(/device/i);
    });

    it("interpolates 'container' into 409 copy", () => {
      const result = msg({ status: 409 }, false, true, "container");
      expect(result).toMatch(/container/i);
      expect(result).not.toMatch(/device/i);
    });

    it("interpolates 'container' into fallback copy", () => {
      const result = msg(new Error("boom"), false, true, "container");
      expect(result).toMatch(/container/i);
      expect(result).not.toMatch(/device/i);
    });
  });

  describe("canSubscribe parameter", () => {
    it("returns owner billing copy when canSubscribe is true on cloud 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      expect(msg({ status: 402 })).toMatch(/update your billing plan/i);
    });

    it("returns non-owner billing copy when canSubscribe is false on cloud 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      expect(msg({ status: 402 }, false, false)).toMatch(
        /ask the namespace owner/i,
      );
    });

    it("returns owner subscription copy when canSubscribe is true on cloud subscription-blocked 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      expect(msg({ status: 402 }, true, true)).toMatch(/open billing/i);
    });

    it("returns non-owner subscription copy when canSubscribe is false on cloud subscription-blocked 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      expect(msg({ status: 402 }, true, false)).toMatch(
        /ask the namespace owner/i,
      );
    });
  });
});
