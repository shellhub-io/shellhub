import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { IAdminNamespace } from "@admin/interfaces/INamespace";

describe("Namespaces Pinia Store", () => {
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const namespaces = [
    {
      name: "namespace1",
      owner: "user1",
      members: [
        { id: "1", role: "operator" as const },
        { id: "2", role: "observer" as const },
        { id: "3", role: "administrator" as const },
      ],
      tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484713",
      created_at: "2023-01-01T00:00:00.000Z",
      devices_accepted_count: 0,
      devices_pending_count: 0,
      devices_rejected_count: 0,
      max_devices: 10,
      settings: {
        session_record: true,
      },
      billing: undefined,
    },
    {
      name: "namespace2",
      owner: "user1",
      members: [
        { id: "4", role: "observer" as const },
        { id: "5", role: "operator" as const },
      ],
      tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484714",
      created_at: "2023-01-01T00:00:00.000Z",
      devices_accepted_count: 1,
      devices_pending_count: 0,
      devices_rejected_count: 0,
      max_devices: 10,
      settings: {
        session_record: false,
      },
      billing: undefined,
    },
    {
      name: "namespace3",
      owner: "user1",
      members: [
        { id: "6", role: "administrator" as const },
        { id: "7", role: "observer" as const },
        { id: "8", role: "operator" as const },
      ],
      tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484715",
      created_at: "2023-01-01T00:00:00.000Z",
      devices_accepted_count: 1,
      devices_pending_count: 1,
      devices_rejected_count: 0,
      max_devices: 10,
      settings: {
        session_record: true,
      },
      billing: undefined,
    },
    {
      name: "namespace4",
      owner: "user1",
      members: [
        { id: "9", role: "operator" as const },
        { id: "10", role: "observer" as const },
      ],
      tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484716",
      created_at: "2023-01-01T00:00:00.000Z",
      devices_accepted_count: 1,
      devices_pending_count: 1,
      devices_rejected_count: 1,
      max_devices: 10,
      settings: {
        session_record: false,
      },
      billing: undefined,
    },
  ];

  const namespace = {
    name: "namespace3",
    owner: "user1",
    members: [
      { id: "6", role: "administrator" as const },
      { id: "7", role: "observer" as const },
      { id: "8", role: "operator" as const },
    ],
    tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484715",
    created_at: "2023-01-01T00:00:00.000Z",
    devices_accepted_count: 1,
    devices_pending_count: 1,
    devices_rejected_count: 0,
    max_devices: 10,
    settings: {
      session_record: true,
    },
    billing: undefined,
  };

  const numberNamespaces = 4;

  beforeEach(() => {
    setActivePinia(createPinia());
    namespacesStore = useNamespacesStore();
  });

  it("returns default state", () => {
    expect(namespacesStore.list).toEqual([]);
    expect(namespacesStore.getNamespace).toEqual({});
  });

  it("sets namespaces and total count", () => {
    namespacesStore.setNamespaces({
      data: namespaces,
      headers: { "x-total-count": numberNamespaces },
    });

    expect(namespacesStore.list).toEqual(namespaces);
    expect(namespacesStore.getnumberOfNamespaces).toEqual(numberNamespaces);
  });

  it("sets a single namespace", () => {
    namespacesStore.setNamespace({ data: namespace });
    expect(namespacesStore.getNamespace).toEqual(namespace);
  });

  it("sets page and perPage", () => {
    namespacesStore.setPageAndPerPage({ perPage: 10, page: 1 });
    expect(namespacesStore.getPerPage).toEqual(10);
    expect(namespacesStore.getPage).toEqual(1);
  });

  it("clears namespaces list", () => {
    namespacesStore.namespaces = namespaces as IAdminNamespace[];
    namespacesStore.clearListNamespaces();
    expect(namespacesStore.list).toEqual([]);
  });
});
