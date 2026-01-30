import { INamespace } from "@/interfaces/INamespace";

/**
 * Mock namespace members for testing.
 */
export const mockMember = {
  id: "user-1",
  role: "owner" as const,
  email: "user1@example.com",
  status: "accepted" as const,
  added_at: "2026-01-30T00:00:00Z",
  expires_at: "2027-01-30T00:00:00Z",
};

export const mockMembers = [
  mockMember,
  { ...mockMember, id: "user-2", email: "user2@example.com", role: "administrator" as const },
  { ...mockMember, id: "user-3", email: "user3@example.com", role: "operator" as const },
  { ...mockMember, id: "user-4", email: "user4@example.com", role: "observer" as const },
];

/**
 * Mock namespace data for testing.
 * Provides a complete namespace object with all required fields.
 */
export const mockNamespace: INamespace = {
  billing: null,
  name: "test",
  owner: "test",
  tenant_id: "fake-tenant-data",
  members: mockMembers,
  settings: {
    session_record: true,
    connection_announcement: "",
  },
  max_devices: 3,
  devices_accepted_count: 0,
  devices_rejected_count: 0,
  devices_pending_count: 0,
  created_at: "",
  type: "personal",
};

export const mockNamespaceWithBilling: INamespace = {
  ...mockNamespace,
  billing: {
    customer_id: "cust_123",
    subscription_id: "sub_123",
    status: "active",
    current_period_end: 1707856000,
    active: true,
  },
};
