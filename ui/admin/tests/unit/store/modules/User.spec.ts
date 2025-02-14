import { describe, expect, it } from "vitest";
import { btoa } from "node:buffer";
import { store } from "../../../../src/store";

describe("users", () => {
  const pagination = {
    page: 1,
    perPage: 10,
  };
  const users = [
    {
      id: "xxxxxxxx",
      name: "user",
      email: "user@email.com",
      username: "username",
      password: "hash_password",
    },
    {
      id: "xxxxxxx2",
      name: "user2",
      email: "user2@email.com",
      username: "username2",
      password: "hash_password2",
    },
    {
      id: "xxxxxxx3",
      name: "user3",
      email: "user3@email.com",
      username: "username3",
      password: "hash_password3",
    },
  ];

  const userData = {
    user: {
      id: "xxxxxxxx",
      name: "user",
      email: "user@email.com",
      username: "username",
      password: "hash_password",
    },
    namespacesOwned: 3,
  };

  const numberUsers = 3;

  const stringToSearch = "user";
  const filterToEncodeBase64 = {
    type: "property",
    params: { name: "name", operator: "contains", value: stringToSearch },
  };
  const encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));

  it("Return firewall default variables", () => {
    expect(store.getters["users/users"]).toEqual([]);
    expect(store.getters["users/user"]).toEqual({});
    expect(store.getters["users/page"]).toEqual(0);
    expect(store.getters["users/perPage"]).toEqual(0);
  });

  it("Verify initial state change for setUsers mutation", () => {
    store.commit("users/setUsers", { data: users, headers: { "x-total-count": numberUsers } });
    expect(store.getters["users/users"]).toEqual(users);
    expect(store.getters["users/numberUsers"]).toEqual(numberUsers);
  });

  it("Verify initial state change for setUser mutation", () => {
    store.commit("users/setUser", { data: userData });
    expect(store.getters["users/user"]).toEqual(userData.user);
    expect(store.getters["users/ownedNamespaces"]).toEqual(userData.namespacesOwned);
  });

  it("Verify initial state change for setPageAndPerPage mutation", () => {
    store.commit("users/setPageAndPerPage", pagination);
    expect(store.getters["users/page"]).toEqual(pagination.page);
    expect(store.getters["users/perPage"]).toEqual(pagination.perPage);
  });

  it("Verify initial state change for setUserFilter mutation", () => {
    store.commit("users/setUserFilter", encodedFilter);
    expect(store.getters["users/filter"]).toEqual(encodedFilter);
  });

  it("Verify initial state change for clearListUsers mutation", () => {
    store.commit("users/clearListUsers");
    expect(store.getters["users/users"]).toEqual([]);
    expect(store.getters["users/numberUsers"]).toEqual(0);
  });
});
