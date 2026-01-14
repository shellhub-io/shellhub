import { IAdminNamespace } from "@admin/interfaces/INamespace";

export const mockNamespace: IAdminNamespace = {
  name: "test-namespace",
  owner: "owner-123",
  tenant_id: "tenant-123",
  members: [
    {
      id: "user-1",
      role: "owner" as const,
      email: "alice@example.com",
      added_at: "2024-01-01T00:00:00Z",
      expires_at: "0001-01-01T00:00:00Z",
      status: "accepted" as const,
    },
    {
      id: "user-2",
      role: "observer" as const,
      email: "bob@example.com",
      added_at: "2024-01-01T00:00:00Z",
      expires_at: "0001-01-01T00:00:00Z",
      status: "accepted" as const,
    },
  ],
  max_devices: 10,
  devices_accepted_count: 3,
  devices_pending_count: 0,
  devices_rejected_count: 0,
  created_at: "2024-01-01T00:00:00Z",
  billing: undefined,
  settings: { session_record: true, connection_announcement: "Welcome!" },
  type: "team" as const,
};

export const mockNamespaces: IAdminNamespace[] = [
  { ...mockNamespace, name: "namespace-one", tenant_id: "tenant-1" },
  { ...mockNamespace, name: "namespace-two", tenant_id: "tenant-2", devices_accepted_count: 5 },
  { ...mockNamespace, name: "namespace-three", tenant_id: "tenant-3", devices_accepted_count: 0 },
];
