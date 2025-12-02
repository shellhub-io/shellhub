import { BasicRole } from "@/interfaces/INamespace";

export interface IInvitation {
  status: "pending" | "expired";
  role: BasicRole;
  invited_by: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
  status_updated_at: string;
  namespace: {
    tenant_id: string;
    name: string;
  };
  user: {
    id: string;
    email: string;
  }
}

export interface IInviteMemberPayload {
  email: string;
  role: BasicRole;
  tenant_id: string;
}
