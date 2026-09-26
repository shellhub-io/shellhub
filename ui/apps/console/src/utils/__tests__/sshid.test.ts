import { describe, it, expect, beforeEach } from "vitest";
import { buildSshid, parseSshid, sshUrl } from "../sshid";

describe("buildSshid", () => {
  beforeEach(() => {
    Object.defineProperty(window, "location", {
      value: { hostname: "shellhub.example.com" },
      writable: true,
    });
  });

  it("returns namespace.deviceName@hostname", () => {
    expect(buildSshid("myns", "mydevice")).toBe(
      "myns.mydevice@shellhub.example.com",
    );
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

describe("parseSshid", () => {
  it("splits an SSHID into namespace, device and host", () => {
    expect(parseSshid("dev.rustagent@localhost")).toEqual({
      namespace: "dev",
      device: "rustagent",
      host: "localhost",
    });
  });

  it("keeps the dots of a device name after the namespace", () => {
    expect(parseSshid("acme.web.eu-1@shellhub.example.com")).toEqual({
      namespace: "acme",
      device: "web.eu-1",
      host: "shellhub.example.com",
    });
  });

  it("reads a target without a namespace as the device alone", () => {
    expect(parseSshid("rustagent@localhost")).toEqual({
      namespace: "",
      device: "rustagent",
      host: "localhost",
    });
  });

  it("returns null for a bare device uid, which has no host", () => {
    expect(parseSshid("3f1c9a0e7b2d")).toBeNull();
  });
});

describe("sshUrl", () => {
  it("puts the login and the SSHID target in the userinfo, encoding its @", () => {
    expect(sshUrl("dev.rustagent@localhost", "deploy")).toBe(
      "ssh://deploy%40dev.rustagent@localhost",
    );
  });

  it("returns null for a bare device uid, which has no host", () => {
    expect(sshUrl("3f1c9a0e7b2d", "root")).toBeNull();
  });
});
