import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { btoa } from "node:buffer";
import useUsersStore from "@admin/store/modules/users";

describe("Users Pinia Store", () => {
  let usersStore: ReturnType<typeof useUsersStore>;

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

  beforeEach(() => {
    setActivePinia(createPinia());
    usersStore = useUsersStore();
  });

  it("returns users default state", () => {
    expect(usersStore.getUsers).toEqual([]);
    expect(usersStore.getUser).toEqual({});
    expect(usersStore.getPage).toEqual(0);
    expect(usersStore.getPerPage).toEqual(0);
  });

  it("sets users and total count", () => {
    usersStore.users = users;
    usersStore.numberUsers = numberUsers;

    expect(usersStore.getUsers).toEqual(users);
    expect(usersStore.getNumberUsers).toEqual(numberUsers);
  });

  it("sets a single user and namespaces owned", () => {
    usersStore.user = userData.user;
    usersStore.ownedNamespaces = userData.namespacesOwned;

    expect(usersStore.getUser).toEqual(userData.user);
    expect(usersStore.getOwnedNamespaces).toEqual(userData.namespacesOwned);
  });

  it("sets pagination", () => {
    usersStore.page = pagination.page;
    usersStore.perPage = pagination.perPage;

    expect(usersStore.getPage).toEqual(pagination.page);
    expect(usersStore.getPerPage).toEqual(pagination.perPage);
  });

  it("sets filter", () => {
    usersStore.filter = encodedFilter;

    expect(usersStore.getFilter).toEqual(encodedFilter);
  });

  it("clears users list", () => {
    usersStore.users = users;
    usersStore.numberUsers = numberUsers;

    usersStore.clearListUsers();

    expect(usersStore.getUsers).toEqual([]);
    expect(usersStore.getNumberUsers).toEqual(0);
  });
});
