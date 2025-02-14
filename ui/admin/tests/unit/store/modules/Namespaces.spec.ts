import { describe, expect, it } from "vitest";
import { store } from "../../../../src/store";

describe("Namespace", () => {
  const namespaces = [
    {
      name: "namespace1",
      owner: "user1",
      members: [{ name: "user3" }, { name: "user4" }, { name: "user5" }],
      tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484713",
    },
    {
      name: "namespace2",
      owner: "user1",
      members: [{ name: "user3" }, { name: "user4" }],
      tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484714",
    },
    {
      name: "namespace3",
      owner: "user1",
      members: [{ name: "user6" }, { name: "user 7" }, { name: "user 8" }],
      tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484715",
    },
    {
      name: "namespace4",
      owner: "user1",
      members: [{ name: "user6" }, { name: "user7" }],
      tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484716",
    },
  ];

  const numberNamespaces = 4;

  const namespace = {
    name: "namespace3",
    owner: "user1",
    members: [{ name: "user6" }, { name: "user7" }, { name: "user8" }],
    tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484715",
  };

  it("Returns namespaces default variables", () => {
    expect(store.getters["namespaces/list"]).toEqual([]);
    expect(store.getters["namespaces/get"]).toEqual({});
  });

  it("Verify initial state change for setNamespaces mutation", () => {
    store.commit("namespaces/setNamespaces", {
      data: namespaces,
      headers: { "x-total-count": numberNamespaces },
    });
    expect(store.getters["namespaces/list"]).toEqual(namespaces);
  });
  it("Verify initial state change for setNamespace mutation", () => {
    store.commit("namespaces/setNamespace", { data: namespace });
    expect(store.getters["namespaces/get"]).toEqual(namespace);
  });

  it("Verify initial state change for setPageAndPerPage mutation", () => {
    store.commit("namespaces/setPageAndPerPage", {
      perPage: 10,
      page: 1,
    });
    expect(store.getters["namespaces/perPage"]).toEqual(10);
    expect(store.getters["namespaces/page"]).toEqual(1);
  });

  it("Verify initial state change for clearListNamespaces mutation", () => {
    store.commit("namespaces/clearListNamespaces");
    expect(store.getters["namespaces/list"]).toEqual([]);
  });
});
