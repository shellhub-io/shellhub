import { describe, it, expect } from "vitest";
import { orderInvitationsByCreatedAt } from "@/utils/invitations";
import { IInvitation } from "@/interfaces/IInvitation";

const makeInvitation = (created_at: string): IInvitation => ({
  status: "pending",
  role: "administrator",
  invited_by: "admin@example.com",
  expires_at: "2026-12-31T00:00:00Z",
  created_at,
  updated_at: created_at,
  status_updated_at: created_at,
  namespace: { tenant_id: "tenant-1", name: "ns" },
  user: { id: "user-1", email: "user@example.com" },
});

describe("orderInvitationsByCreatedAt", () => {
  it("returns invitations sorted newest first", () => {
    const invitations = [
      makeInvitation("2026-01-01T00:00:00Z"),
      makeInvitation("2026-03-01T00:00:00Z"),
      makeInvitation("2026-02-01T00:00:00Z"),
    ];

    const sorted = orderInvitationsByCreatedAt(invitations);

    expect(sorted.map((i) => i.created_at)).toEqual([
      "2026-03-01T00:00:00Z",
      "2026-02-01T00:00:00Z",
      "2026-01-01T00:00:00Z",
    ]);
  });

  it("does not mutate the original array", () => {
    const invitations = [
      makeInvitation("2026-01-01T00:00:00Z"),
      makeInvitation("2026-03-01T00:00:00Z"),
    ];
    const original = [...invitations];

    orderInvitationsByCreatedAt(invitations);

    expect(invitations).toEqual(original);
  });
});
