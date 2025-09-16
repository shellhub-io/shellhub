import { describe, expect, it } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import { IAdminUser } from "@admin/interfaces/IUser";

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

const numberUsers = 3;

describe("Users Pinia Store", () => {
  setActivePinia(createPinia());
  const usersStore = useUsersStore();

  it("returns users default state", () => {
    expect(usersStore.users).toEqual([]);
  });

  it("sets users and total count", () => {
    usersStore.users = users as IAdminUser[];
    usersStore.usersCount = numberUsers;

    expect(usersStore.users).toEqual(users);
    expect(usersStore.usersCount).toEqual(numberUsers);
  });
});
