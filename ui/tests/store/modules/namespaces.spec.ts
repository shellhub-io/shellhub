import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { namespacesApi } from "@/api/http";
import { INamespace, INamespaceMember } from "@/interfaces/INamespace";
import useNamespacesStore from "@/store/modules/namespaces";

const mockNamespaceMemberBase: INamespaceMember = {
  id: "507f1f77bcf86cd799439011",
  role: "administrator",
  email: "admin@example.com",
  expires_at: "2026-12-31T23:59:59.000Z",
  added_at: "2025-12-31T23:59:59.000Z",
};

const mockNamespaceBase: INamespace = {
  name: "examplespace",
  owner: "507f1f77bcf86cd799439011",
  tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
  members: [mockNamespaceMemberBase],
  settings: {
    session_record: true,
    connection_announcement: "",
  },
  max_devices: 3,
  devices_accepted_count: 0,
  devices_pending_count: 0,
  devices_rejected_count: 0,
  created_at: "2025-05-01T00:00:00.000Z",
  billing: null,
  type: "team",
};

describe("Namespaces Store", () => {
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let mockNamespacesApi: MockAdapter;

  beforeEach(() => {
    localStorage.clear();
    setActivePinia(createPinia());
    namespacesStore = useNamespacesStore();
    mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
  });

  afterEach(() => {
    mockNamespacesApi.reset();
  });

  describe("Initial State", () => {
    it("should have empty current namespace", () => {
      expect(namespacesStore.currentNamespace).toEqual({});
    });

    it("should have empty namespace list", () => {
      expect(namespacesStore.namespaceList).toEqual([]);
    });

    it("should have undefined user status", () => {
      expect(namespacesStore.userStatus).toBeUndefined();
    });

    it("should have hasNamespaces computed as false when list is empty", () => {
      expect(namespacesStore.hasNamespaces).toBe(false);
    });

    it("should have hasNamespaces computed as true when list has namespaces", () => {
      namespacesStore.namespaceList = [mockNamespaceBase];
      expect(namespacesStore.hasNamespaces).toBe(true);
    });
  });

  describe("fetchNamespaceList", () => {
    it("should fetch namespace list successfully with default pagination", async () => {
      const namespaceList = [mockNamespaceBase];

      mockNamespacesApi
        .onGet("http://localhost:3000/api/namespaces?page=1&per_page=10")
        .reply(200, namespaceList);

      await expect(namespacesStore.fetchNamespaceList()).resolves.not.toThrow();

      expect(namespacesStore.namespaceList).toEqual(namespaceList);
    });

    it("should fetch namespace list successfully with custom pagination", async () => {
      const namespaceList = [mockNamespaceBase];

      mockNamespacesApi
        .onGet("http://localhost:3000/api/namespaces?page=2&per_page=20")
        .reply(200, namespaceList);

      await expect(namespacesStore.fetchNamespaceList({ page: 2, perPage: 20 })).resolves.not.toThrow();

      expect(namespacesStore.namespaceList).toEqual(namespaceList);
    });

    it("should fetch namespace list successfully with filter", async () => {
      const namespaceList = [mockNamespaceBase];
      const filter = "example";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces?filter=${filter}&page=1&per_page=10`)
        .reply(200, namespaceList);

      await expect(namespacesStore.fetchNamespaceList({ page: 1, perPage: 10, filter })).resolves.not.toThrow();

      expect(namespacesStore.namespaceList).toEqual(namespaceList);
    });

    it("should handle not found error when fetching namespace list", async () => {
      mockNamespacesApi
        .onGet("http://localhost:3000/api/namespaces?page=1&per_page=10")
        .reply(404, { message: "No namespaces found" });

      await expect(namespacesStore.fetchNamespaceList()).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when fetching namespace list", async () => {
      mockNamespacesApi
        .onGet("http://localhost:3000/api/namespaces?page=1&per_page=10")
        .reply(500);

      await expect(namespacesStore.fetchNamespaceList()).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when fetching namespace list", async () => {
      mockNamespacesApi
        .onGet("http://localhost:3000/api/namespaces?page=1&per_page=10")
        .networkError();

      await expect(namespacesStore.fetchNamespaceList()).rejects.toThrow("Network Error");
    });
  });

  describe("fetchNamespace", () => {
    it("should fetch namespace by id successfully", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId}`)
        .reply(200, mockNamespaceBase);

      await expect(namespacesStore.fetchNamespace(tenantId)).resolves.not.toThrow();

      expect(namespacesStore.currentNamespace).toEqual(mockNamespaceBase);
    });

    it("should handle not found error when fetching namespace", async () => {
      const tenantId = "invalid-id";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId}`)
        .reply(404, { message: "Namespace not found" });

      await expect(namespacesStore.fetchNamespace(tenantId)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when fetching namespace", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId}`)
        .reply(500);

      await expect(namespacesStore.fetchNamespace(tenantId)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when fetching namespace", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId}`)
        .networkError();

      await expect(namespacesStore.fetchNamespace(tenantId)).rejects.toThrow("Network Error");
    });
  });

  describe("createNamespace", () => {
    it("should create namespace successfully and return tenant id", async () => {
      const namespaceName = "newnamespace";
      const tenantId = "new-tenant-id";

      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces")
        .reply(200, { tenant_id: tenantId });

      const result = await namespacesStore.createNamespace(namespaceName);

      expect(result).toBe(tenantId);
    });

    it("should handle validation error when creating namespace", async () => {
      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces")
        .reply(400, { message: "Invalid namespace name" });

      await expect(namespacesStore.createNamespace("invalid name")).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle server error when creating namespace", async () => {
      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces")
        .reply(500);

      await expect(namespacesStore.createNamespace("newnamespace")).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when creating namespace", async () => {
      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces")
        .networkError();

      await expect(namespacesStore.createNamespace("newnamespace")).rejects.toThrow("Network Error");
    });
  });

  describe("editNamespace", () => {
    it("should edit namespace successfully and update current namespace", async () => {
      const editData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        name: "updatedspace",
        settings: {
          session_record: false,
          connection_announcement: "Welcome",
        },
      };

      const updatedNamespace = {
        ...mockNamespaceBase,
        name: editData.name,
        settings: editData.settings,
      };

      mockNamespacesApi
        .onPut(`http://localhost:3000/api/namespaces/${editData.tenant_id}`)
        .reply(200, updatedNamespace);

      await expect(namespacesStore.editNamespace(editData)).resolves.not.toThrow();

      expect(namespacesStore.currentNamespace).toEqual(updatedNamespace);
    });

    it("should handle not found error when editing namespace", async () => {
      const editData = {
        tenant_id: "invalid-id",
        name: "updatedspace",
      };

      mockNamespacesApi
        .onPut(`http://localhost:3000/api/namespaces/${editData.tenant_id}`)
        .reply(404, { message: "Namespace not found" });

      await expect(namespacesStore.editNamespace(editData)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when editing namespace", async () => {
      const editData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        name: "updatedspace",
      };

      mockNamespacesApi
        .onPut(`http://localhost:3000/api/namespaces/${editData.tenant_id}`)
        .reply(500);

      await expect(namespacesStore.editNamespace(editData)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when editing namespace", async () => {
      const editData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        name: "updatedspace",
      };

      mockNamespacesApi
        .onPut(`http://localhost:3000/api/namespaces/${editData.tenant_id}`)
        .networkError();

      await expect(namespacesStore.editNamespace(editData)).rejects.toThrow("Network Error");
    });
  });

  describe("deleteNamespace", () => {
    it("should delete namespace successfully and reset state", async () => {
      namespacesStore.currentNamespace = mockNamespaceBase;
      namespacesStore.namespaceList = [mockNamespaceBase];

      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${tenantId}`)
        .reply(200);

      await expect(namespacesStore.deleteNamespace(tenantId)).resolves.not.toThrow();

      expect(namespacesStore.currentNamespace).toEqual({});
      expect(namespacesStore.namespaceList).toEqual([]);
    });

    it("should handle not found error when deleting namespace", async () => {
      const tenantId = "invalid-id";

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${tenantId}`)
        .reply(404, { message: "Namespace not found" });

      await expect(namespacesStore.deleteNamespace(tenantId)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when deleting namespace", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${tenantId}`)
        .reply(500);

      await expect(namespacesStore.deleteNamespace(tenantId)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when deleting namespace", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${tenantId}`)
        .networkError();

      await expect(namespacesStore.deleteNamespace(tenantId)).rejects.toThrow("Network Error");
    });
  });

  describe("leaveNamespace", () => {
    it("should leave namespace successfully and update localStorage with new tenant", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";
      const leaveResponse = {
        token: "new-token",
        tenant: "new-tenant-id",
        role: "member",
      };

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${tenantId}/members`)
        .reply(200, leaveResponse);

      await expect(namespacesStore.leaveNamespace(tenantId)).resolves.not.toThrow();

      expect(localStorage.getItem("token")).toBe("new-token");
      expect(localStorage.getItem("tenant")).toBe("new-tenant-id");
      expect(localStorage.getItem("role")).toBe("member");
    });

    it("should leave namespace successfully and update localStorage without new tenant", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";
      const leaveResponse = {
        token: "new-token",
      };

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${tenantId}/members`)
        .reply(200, leaveResponse);

      await expect(namespacesStore.leaveNamespace(tenantId)).resolves.not.toThrow();

      expect(localStorage.getItem("token")).toBe("new-token");
    });

    it("should handle forbidden error when leaving namespace", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${tenantId}/members`)
        .reply(403, { message: "Cannot leave namespace as owner" });

      await expect(namespacesStore.leaveNamespace(tenantId)).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should handle server error when leaving namespace", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${tenantId}/members`)
        .reply(500);

      await expect(namespacesStore.leaveNamespace(tenantId)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when leaving namespace", async () => {
      const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${tenantId}/members`)
        .networkError();

      await expect(namespacesStore.leaveNamespace(tenantId)).rejects.toThrow("Network Error");
    });
  });

  describe("updateNamespaceMember", () => {
    it("should update namespace member successfully", async () => {
      const updateData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        user_id: "507f1f77bcf86cd799439011",
        role: "observer" as const,
      };

      mockNamespacesApi
        .onPatch(`http://localhost:3000/api/namespaces/${updateData.tenant_id}/members/${updateData.user_id}`)
        .reply(200);

      await expect(namespacesStore.updateNamespaceMember(updateData)).resolves.not.toThrow();
    });

    it("should handle not found error when updating namespace member", async () => {
      const updateData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        user_id: "invalid-user-id",
        role: "observer" as const,
      };

      mockNamespacesApi
        .onPatch(`http://localhost:3000/api/namespaces/${updateData.tenant_id}/members/${updateData.user_id}`)
        .reply(404, { message: "Member not found" });

      await expect(namespacesStore.updateNamespaceMember(updateData)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when updating namespace member", async () => {
      const updateData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        user_id: "507f1f77bcf86cd799439011",
        role: "observer" as const,
      };

      mockNamespacesApi
        .onPatch(`http://localhost:3000/api/namespaces/${updateData.tenant_id}/members/${updateData.user_id}`)
        .reply(500);

      await expect(namespacesStore.updateNamespaceMember(updateData)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when updating namespace member", async () => {
      const updateData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        user_id: "507f1f77bcf86cd799439011",
        role: "observer" as const,
      };

      mockNamespacesApi
        .onPatch(`http://localhost:3000/api/namespaces/${updateData.tenant_id}/members/${updateData.user_id}`)
        .networkError();

      await expect(namespacesStore.updateNamespaceMember(updateData)).rejects.toThrow("Network Error");
    });
  });

  describe("removeMemberFromNamespace", () => {
    it("should remove member from namespace successfully", async () => {
      const removeData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        user_id: "507f1f77bcf86cd799439011",
      };

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${removeData.tenant_id}/members/${removeData.user_id}`)
        .reply(200);

      await expect(namespacesStore.removeMemberFromNamespace(removeData)).resolves.not.toThrow();
    });

    it("should handle not found error when removing member from namespace", async () => {
      const removeData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        user_id: "invalid-user-id",
      };

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${removeData.tenant_id}/members/${removeData.user_id}`)
        .reply(404, { message: "Member not found" });

      await expect(namespacesStore.removeMemberFromNamespace(removeData)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when removing member from namespace", async () => {
      const removeData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        user_id: "507f1f77bcf86cd799439011",
      };

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${removeData.tenant_id}/members/${removeData.user_id}`)
        .reply(500);

      await expect(namespacesStore.removeMemberFromNamespace(removeData)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when removing member from namespace", async () => {
      const removeData = {
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        user_id: "507f1f77bcf86cd799439011",
      };

      mockNamespacesApi
        .onDelete(`http://localhost:3000/api/namespaces/${removeData.tenant_id}/members/${removeData.user_id}`)
        .networkError();

      await expect(namespacesStore.removeMemberFromNamespace(removeData)).rejects.toThrow("Network Error");
    });
  });

  describe("lookupUserStatus", () => {
    it("should lookup user status successfully and update state", async () => {
      const lookupData = {
        tenant: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        id: "user-id",
        sig: "signature",
      };
      const statusResponse = { status: "invited" };

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${lookupData.tenant}/members/${lookupData.id}/accept-invite?sig=${lookupData.sig}`)
        .reply(200, statusResponse);

      await expect(namespacesStore.lookupUserStatus(lookupData)).resolves.not.toThrow();

      expect(namespacesStore.userStatus).toBe("invited");
    });

    it("should handle not found error when looking up user status", async () => {
      const lookupData = {
        tenant: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        id: "invalid-user-id",
        sig: "signature",
      };

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${lookupData.tenant}/members/${lookupData.id}/accept-invite?sig=${lookupData.sig}`)
        .reply(404, { message: "Invitation not found" });

      await expect(namespacesStore.lookupUserStatus(lookupData)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when looking up user status", async () => {
      const lookupData = {
        tenant: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        id: "user-id",
        sig: "signature",
      };

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${lookupData.tenant}/members/${lookupData.id}/accept-invite?sig=${lookupData.sig}`)
        .reply(500);

      await expect(namespacesStore.lookupUserStatus(lookupData)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when looking up user status", async () => {
      const lookupData = {
        tenant: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        id: "user-id",
        sig: "signature",
      };

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${lookupData.tenant}/members/${lookupData.id}/accept-invite?sig=${lookupData.sig}`)
        .networkError();

      await expect(namespacesStore.lookupUserStatus(lookupData)).rejects.toThrow("Network Error");
    });
  });

  describe("switchNamespace", () => {
    it("should switch namespace successfully and update localStorage", async () => {
      const tenantId = "new-tenant-id";
      const switchResponse = {
        token: "new-token",
        role: "admin",
      };

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/auth/token/${tenantId}`)
        .reply(200, switchResponse);

      await expect(namespacesStore.switchNamespace(tenantId)).resolves.not.toThrow();

      expect(localStorage.getItem("token")).toBe("new-token");
      expect(localStorage.getItem("tenant")).toBe(tenantId);
      expect(localStorage.getItem("role")).toBe("admin");
    });

    it("should handle forbidden error when switching namespace", async () => {
      const tenantId = "unauthorized-tenant-id";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/auth/token/${tenantId}`)
        .reply(403, { message: "Access denied to namespace" });

      await expect(namespacesStore.switchNamespace(tenantId)).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should handle server error when switching namespace", async () => {
      const tenantId = "new-tenant-id";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/auth/token/${tenantId}`)
        .reply(500);

      await expect(namespacesStore.switchNamespace(tenantId)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when switching namespace", async () => {
      const tenantId = "new-tenant-id";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/auth/token/${tenantId}`)
        .networkError();

      await expect(namespacesStore.switchNamespace(tenantId)).rejects.toThrow("Network Error");
    });
  });

  describe("reset", () => {
    it("should reset store state to initial values", () => {
      namespacesStore.currentNamespace = mockNamespaceBase;
      namespacesStore.namespaceList = [mockNamespaceBase];
      namespacesStore.userStatus = "invited";

      namespacesStore.reset();

      expect(namespacesStore.currentNamespace).toEqual({});
      expect(namespacesStore.namespaceList).toEqual([]);
      expect(namespacesStore.userStatus).toBeUndefined();
    });
  });
});
