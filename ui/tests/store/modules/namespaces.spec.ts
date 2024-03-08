import { describe, expect, it } from "vitest";
import { store } from "@/store";

describe("Namespaces Store", () => {
  const namespaceData = {
    data: {
      name: "examplespace",
      owner: "507f1f77bcf86cd799439011",
      tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
      members: [
        {
          id: "507f1f77bcf86cd799439011",
          role: "administrator",
        },
      ],
      settings: {
        session_record: true,
        connection_announcement: "",
      },
      max_devices: 3,
      device_count: 0,
      created_at: "2020-05-01T00:00:00.000Z",
      billing: null,
    },
  };

  it("Returns namespaces with default variables", () => {
    expect(store.getters["namespaces/list"]).toEqual([]);
    expect(store.getters["namespaces/get"]).toEqual({});
    expect(store.getters["namespaces/getNumberNamespaces"]).toEqual(0);
    expect(store.getters["namespaces/owner"]).toEqual(false);
    expect(store.getters["namespaces/billing"]).toEqual({});
  });

  it("Commits setNamespaces mutation", async () => {
    const mockData = { data: [], headers: { "x-total-count": "0" } };
    store.commit("namespaces/setNamespaces", mockData);
    expect(store.getters["namespaces/list"]).toEqual(mockData.data);
    expect(store.getters["namespaces/getNumberNamespaces"]).toEqual(parseInt(mockData.headers["x-total-count"], 10));
  });

  it("Commits setNamespace mutation", async () => {
    const mockData = { data: {} };
    store.commit("namespaces/setNamespace", mockData);
    expect(store.getters["namespaces/get"]).toEqual(mockData.data);
  });

  it("putNamespace mutation", async () => {
    store.commit("namespaces/setNamespace", namespaceData);
    expect(store.getters["namespaces/get"]).toEqual(namespaceData.data);
  });
  // Add more tests for other actions and mutations
});
