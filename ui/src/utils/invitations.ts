import { IInvitation } from "@/interfaces/IInvitation";

export const getInvitationStatusFilter = (status: IInvitation["status"]) => {
  const filter = [{
    type: "property",
    params: { name: "status", operator: "eq", value: status },
  }];

  return Buffer.from(JSON.stringify(filter)).toString("base64");
};

export const orderInvitationsByCreatedAt = (invitations: IInvitation[]) => {
  return [...invitations].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
};

export const isInvitationExpired = (expiresAt: IInvitation["expires_at"]): boolean => {
  if (!expiresAt) return false;
  return new Date(expiresAt).getTime() < Date.now();
};
