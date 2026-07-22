import { describe, it, expect } from "vitest";
import { parseSourceIp, sourceIpKind } from "../sourceIp";

describe("parseSourceIp", () => {
  it("normalizes a bare IPv4 to a /32 host route", () => {
    const r = parseSourceIp("10.0.0.5");
    expect(r).toMatchObject({ status: "host", value: "10.0.0.5/32" });
  });

  it("normalizes a bare IPv6 to a /128 host route", () => {
    expect(parseSourceIp("2001:db8::1")).toMatchObject({
      status: "host",
      value: "2001:db8::1/128",
    });
  });

  it("accepts and classifies a private CIDR", () => {
    const r = parseSourceIp("10.0.0.0/8");
    expect(r.status).toBe("valid");
    if (r.status === "valid") expect(r.label).toMatch(/private/i);
  });

  it("classifies a public CIDR", () => {
    const r = parseSourceIp("203.0.113.0/24");
    expect(r.status).toBe("valid");
    if (r.status === "valid") expect(r.label).toMatch(/public/i);
  });

  it("flags 0.0.0.0/0 as any", () => {
    expect(parseSourceIp("0.0.0.0/0")).toMatchObject({
      status: "any",
      value: "0.0.0.0/0",
    });
  });

  it("rejects an out-of-range octet", () => {
    expect(parseSourceIp("10.0.0.999").status).toBe("invalid");
  });

  it("rejects an out-of-range prefix", () => {
    expect(parseSourceIp("10.0.0.0/40").status).toBe("invalid");
  });

  it("treats still-being-typed input as incomplete, not invalid", () => {
    expect(parseSourceIp("10.0.").status).toBe("incomplete");
  });

  it("rejects clearly bogus input", () => {
    expect(parseSourceIp("hello").status).toBe("invalid");
  });

  it("is empty for blank input", () => {
    expect(parseSourceIp("   ").status).toBe("empty");
  });
});

describe("sourceIpKind", () => {
  it.each([
    ["10.0.0.5/32", "host"],
    ["10.0.0.0/8", "private"],
    ["192.168.0.0/16", "private"],
    ["203.0.113.0/24", "public"],
    ["0.0.0.0/0", "any"],
    ["2001:db8::/32", "ipv6"],
  ])("labels %s as %s", (cidr, kind) => {
    expect(sourceIpKind(cidr)).toBe(kind);
  });
});
