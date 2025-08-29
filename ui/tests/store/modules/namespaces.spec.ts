import { describe, expect, it, beforeEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { namespacesApi } from "@/api/http";
import { INamespace, INamespaceMember } from "@/interfaces/INamespace";
import useNamespacesStore from "@/store/modules/namespaces";

const namespaceData: INamespace = {
  name: "examplespace",
  owner: "507f1f77bcf86cd799439011",
  tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
  members: [
    {
      id: "507f1f77bcf86cd799439011",
      role: "administrator",
    },
  ] as INamespaceMember[],
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
};

describe("Namespaces Store", () => {
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  beforeEach(() => {
    namespacesStore.currentNamespace = {} as INamespace;
    namespacesStore.namespaceList = [];
    namespacesStore.userStatus = undefined;
  });

  it("should have initial state values", () => {
    expect(namespacesStore.currentNamespace).toEqual({});
    expect(namespacesStore.namespaceList).toEqual([]);
    expect(namespacesStore.userStatus).toBeUndefined();
  });

  it("should fetch namespace list successfully", async () => {
    const namespacesData = [namespaceData];

    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces?page=1&per_page=10").reply(200, namespacesData);

    await namespacesStore.fetchNamespaceList({ page: 1, perPage: 10 });

    expect(namespacesStore.namespaceList).toEqual(namespacesData);
  });

  it("should fetch namespace by id successfully", async () => {
    const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

    mockNamespacesApi.onGet(`http://localhost:3000/api/namespaces/${tenantId}`).reply(200, namespaceData);

    await namespacesStore.fetchNamespace(tenantId);

    expect(namespacesStore.currentNamespace).toEqual(namespaceData);
  });

  it("should create namespace successfully", async () => {
    const tenantId = "new-tenant-id";

    mockNamespacesApi.onPost("http://localhost:3000/api/namespaces").reply(200, { tenant_id: tenantId });

    const result = await namespacesStore.createNamespace("newnamespace");

    expect(result).toEqual(tenantId);
  });

  it("should edit namespace successfully", async () => {
    const editData = {
      tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
      name: "updatedspace",
    };

    const updatedNamespace = { ...namespaceData, name: "updatedspace" };

    mockNamespacesApi.onPut(`http://localhost:3000/api/namespaces/${editData.tenant_id}`).reply(200, updatedNamespace);

    await namespacesStore.editNamespace(editData);

    expect(namespacesStore.currentNamespace).toEqual(updatedNamespace);
  });

  it("should delete namespace successfully", async () => {
    namespacesStore.currentNamespace = namespaceData;
    namespacesStore.namespaceList = [namespaceData];

    const tenantId = "3dd0d1f8-8246-4519-b11a-a3dd33717f65";

    mockNamespacesApi.onDelete(`http://localhost:3000/api/namespaces/${tenantId}`).reply(200);

    await namespacesStore.deleteNamespace(tenantId);

    expect(namespacesStore.currentNamespace).toEqual({});
    expect(namespacesStore.namespaceList).toEqual([]);
  });

  it("should switch namespace successfully", async () => {
    const tenantId = "new-tenant-id";
    const switchResponse = {
      token: "new-token",
      role: "admin",
    };

    mockNamespacesApi.onGet(`http://localhost:3000/api/auth/token/${tenantId}`).reply(200, switchResponse);

    await namespacesStore.switchNamespace(tenantId);

    expect(localStorage.getItem("token")).toEqual("new-token");
    expect(localStorage.getItem("tenant")).toEqual(tenantId);
    expect(localStorage.getItem("role")).toEqual("admin");
  });

  it("should lookup user status successfully", async () => {
    const lookupData = {
      tenant: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
      id: "user-id",
      sig: "signature",
    };
    const statusResponse = { status: "invited" };

    const url = `http://localhost:3000/api/namespaces/${lookupData.tenant}/members/${lookupData.id}/accept-invite`
        + `?sig=${lookupData.sig}`;

    mockNamespacesApi.onGet(url).reply(200, statusResponse);

    await namespacesStore.lookupUserStatus(lookupData);
    expect(namespacesStore.userStatus).toEqual("invited");
  });

  it("should handle fetch namespace list error", async () => {
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces?page=1&per_page=10").reply(500);
    await expect(namespacesStore.fetchNamespaceList()).rejects.toThrow();
  });

  it("should handle fetch namespace error", async () => {
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/invalid-id").reply(404);
    await expect(namespacesStore.fetchNamespace("invalid-id")).rejects.toThrow();
  });

  it("should handle create namespace error", async () => {
    mockNamespacesApi.onPost("http://localhost:3000/api/namespaces").reply(400);
    await expect(namespacesStore.createNamespace("invalidname")).rejects.toThrow();
  });
});
