import type { MembershipInvitation } from "@/client";
import { toBase64Json } from "@/utils/encoding";

/**
 * The status of a membership invitation, taken from the generated model so the two cannot drift.
 */
export type InvitationStatus = MembershipInvitation["status"];

/**
 * Builds the base64-encoded filter query used by the backend's invitation list
 * endpoints. Mirrors the encoding from the legacy Vue UI (ui/src/utils/invitations.ts).
 */
export function invitationStatusFilter(status: InvitationStatus): string {
  const filter = [
    {
      type: "property",
      params: { name: "status", operator: "eq", value: status },
    },
  ];
  return toBase64Json(filter);
}

/**
 * Whether an invitation is past its expiry. A null expiresAt means it does not expire, which is
 * not the same as expired — the distinction decides whether the row offers a resend.
 */
export function isInvitationExpired(expiresAt: string | null): boolean {
  if (!expiresAt) return false;
  return new Date(expiresAt).getTime() < Date.now();
}
