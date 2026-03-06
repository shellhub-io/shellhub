import { describe, it, expect, beforeEach } from "vitest";
import { buildSshid } from "../sshid";

describe("buildSshid", () => {
  beforeEach(() => {
    Object.defineProperty(window, "location", {
      value: { hostname: "shellhub.example.com" },
      writable: true,
    });
  });

  it("returns namespace.deviceName@hostname", () => {
    expect(buildSshid("myns", "mydevice")).toBe("myns.mydevice@shellhub.example.com");
  });

  it("uses window.location.hostname as the server part, not the namespace", () => {
    const result = buildSshid("myns", "mydevice");
    expect(result).not.toContain("@myns");
    expect(result).toContain("@shellhub.example.com");
  });

  it("reflects changes in window.location.hostname", () => {
    Object.defineProperty(window, "location", {
      value: { hostname: "other.host.io" },
      writable: true,
    });
    expect(buildSshid("ns", "dev")).toBe("ns.dev@other.host.io");
  });

  it("handles localhost as hostname", () => {
    Object.defineProperty(window, "location", {
      value: { hostname: "localhost" },
      writable: true,
    });
    expect(buildSshid("dev", "agent")).toBe("dev.agent@localhost");
  });
});
