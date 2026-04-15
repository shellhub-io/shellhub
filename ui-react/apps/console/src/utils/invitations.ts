import type { MembershipInvitation } from "@/client";

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
  return btoa(JSON.stringify(filter));
}

export function isInvitationExpired(expiresAt: string | null): boolean {
  if (!expiresAt) return false;
  return new Date(expiresAt).getTime() < Date.now();
}
