import { IInvitation } from "@/interfaces/IInvitation";

/**
 * Mock invitation data for testing.
 * Provides a complete invitation object with all required fields.
 */
export const mockInvitation: IInvitation = {
  status: "pending",
  role: "operator",
  invited_by: "admin-user",
  expires_at: "2025-12-31T23:59:59Z",
  created_at: "2025-12-01T00:00:00Z",
  updated_at: "2025-12-01T00:00:00Z",
  status_updated_at: "2025-12-01T00:00:00Z",
  namespace: {
    tenant_id: "tenant1",
    name: "Test Namespace",
  },
  user: {
    id: "user1",
    email: "user@example.com",
  },
};

/**
 * Mock invitation with administrator role.
 */
export const mockInvitationAdmin: IInvitation = {
  ...mockInvitation,
  role: "administrator",
  namespace: {
    tenant_id: "tenant2",
    name: "Admin Namespace",
  },
};

/**
 * Mock list of invitations for testing lists.
 */
export const mockInvitations: IInvitation[] = [
  mockInvitation,
  mockInvitationAdmin,
];
