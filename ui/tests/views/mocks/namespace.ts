import { INamespace } from "@/interfaces/INamespace";

/**
 * Mock namespace members for testing.
 */
export const mockMembers = [
  {
    id: "xxxxxxxx",
    role: "owner" as const,
    email: "test@example.com",
    status: "accepted" as const,
    added_at: "2024-01-01T00:00:00Z",
    expires_at: "2025-01-01T00:00:00Z",
  },
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
