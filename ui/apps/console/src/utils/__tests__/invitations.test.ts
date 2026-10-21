import { describe, it, expect, vi, afterEach } from "vitest";
import { invitationStatusFilter, isInvitationExpired } from "../invitations";
import { decodeB64url } from "@/tests/decodeB64url";

afterEach(() => {
  vi.restoreAllMocks();
});

describe("invitationStatusFilter", () => {
  it("decodes to valid JSON array", () => {
    const result = invitationStatusFilter("pending");
    const decoded = decodeB64url(result) as unknown[];
    expect(Array.isArray(decoded)).toBe(true);
    expect(decoded).toHaveLength(1);
  });

  it("encodes a property-type filter with name='status' and operator='eq'", () => {
    const result = invitationStatusFilter("accepted");
    const [filter] = decodeB64url(result) as Array<{
      type: string;
      params: { name: string; operator: string; value: string };
    }>;
    expect(filter.type).toBe("property");
    expect(filter.params.name).toBe("status");
    expect(filter.params.operator).toBe("eq");
  });

  it("encodes the given status as the filter value — pending", () => {
    const result = invitationStatusFilter("pending");
    const [filter] = decodeB64url(result) as Array<{
      type: string;
      params: { name: string; operator: string; value: string };
    }>;
    expect(filter.params.value).toBe("pending");
  });

  it("encodes the given status as the filter value — cancelled", () => {
    const result = invitationStatusFilter("cancelled");
    const [filter] = decodeB64url(result) as Array<{
      type: string;
      params: { name: string; operator: string; value: string };
    }>;
    expect(filter.params.value).toBe("cancelled");
  });

  it("decodes a base64url value that contains '-' (from '+' in standard base64)", () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const result = invitationStatusFilter(">" as any);
    expect(result).not.toMatch(/\+/);
    expect(result).not.toMatch(/=$/);
    const [filter] = decodeB64url(result) as Array<{
      type: string;
      params: { name: string; operator: string; value: string };
    }>;
    expect(filter.params.value).toBe(">");
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
});
