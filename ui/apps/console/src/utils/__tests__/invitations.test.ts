import { describe, it, expect, vi, afterEach } from "vitest";
import { invitationStatusFilter, isInvitationExpired } from "../invitations";

afterEach(() => {
  vi.restoreAllMocks();
});

describe("invitationStatusFilter", () => {
  it("returns a non-empty base64 string", () => {
    const result = invitationStatusFilter("pending");
    expect(typeof result).toBe("string");
    expect(result.length).toBeGreaterThan(0);
  });

  it("decodes to valid JSON array", () => {
    const result = invitationStatusFilter("pending");
    const decoded = JSON.parse(atob(result)) as unknown[];
    expect(Array.isArray(decoded)).toBe(true);
    expect(decoded).toHaveLength(1);
  });

  it("encodes a property-type filter with name='status' and operator='eq'", () => {
    const result = invitationStatusFilter("accepted");
    const [filter] = JSON.parse(atob(result)) as Array<{
      type: string;
      params: { name: string; operator: string; value: string };
    }>;
    expect(filter.type).toBe("property");
    expect(filter.params.name).toBe("status");
    expect(filter.params.operator).toBe("eq");
  });

  it("encodes the given status as the filter value — pending", () => {
    const result = invitationStatusFilter("pending");
    const [filter] = JSON.parse(atob(result)) as Array<{
      type: string;
      params: { name: string; operator: string; value: string };
    }>;
    expect(filter.params.value).toBe("pending");
  });

  it("encodes the given status as the filter value — cancelled", () => {
    const result = invitationStatusFilter("cancelled");
    const [filter] = JSON.parse(atob(result)) as Array<{
      type: string;
      params: { name: string; operator: string; value: string };
    }>;
    expect(filter.params.value).toBe("cancelled");
  });

  it("produces different output for different statuses", () => {
    expect(invitationStatusFilter("pending")).not.toBe(
      invitationStatusFilter("accepted"),
    );
  });
});

describe("isInvitationExpired", () => {
  it("returns false when expiresAt is null", () => {
    expect(isInvitationExpired(null)).toBe(false);
  });

  it("returns true when expiresAt is in the past", () => {
    const pastDate = new Date(Date.now() - 60_000).toISOString();
    expect(isInvitationExpired(pastDate)).toBe(true);
  });

  it("returns false when expiresAt is in the future", () => {
    const futureDate = new Date(Date.now() + 60_000).toISOString();
    expect(isInvitationExpired(futureDate)).toBe(false);
  });

  it("returns true when expiresAt is exactly at epoch 0 (far past)", () => {
    expect(isInvitationExpired("1970-01-01T00:00:00.000Z")).toBe(true);
  });

  it("returns false when expiresAt is far in the future", () => {
    expect(isInvitationExpired("2099-12-31T23:59:59.000Z")).toBe(false);
  });
});
