import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import useUsersStore from "@admin/store/modules/users";
import { IAdminUser, IAdminUserFormData } from "@admin/interfaces/IUser";
import { buildUrl } from "@tests/utils/url";

const mockUserBase: IAdminUser = {
  id: "user-id-123",
  status: "confirmed",
  max_namespaces: 5,
  created_at: "2026-01-01T00:00:00.000Z",
  last_login: "2026-01-01T10:00:00.000Z",
  name: "Admin User",
  username: "admin",
  email: "admin@example.com",
  recovery_email: "recovery@example.com",
  mfa: { enabled: false },
  namespacesOwned: 2,
  preferences: {
    auth_methods: ["local"],
  },
  email_marketing: false,
  admin: true,
};

const mockUserFormData: IAdminUserFormData = {
  name: "New User",
  email: "newuser@example.com",
  username: "newuser",
  password: "password123",
  max_namespaces: 3,
  status: "confirmed",
  admin: false,
};

describe("Admin Users Store", () => {
  let usersStore: ReturnType<typeof useUsersStore>;
  let mockAdminApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    usersStore = useUsersStore();
    mockAdminApi = new MockAdapter(adminApi.getAxios());
  });

  afterEach(() => { mockAdminApi.reset(); });

  describe("Initial State", () => {
    it("should have empty users array", () => {
      expect(usersStore.users).toEqual([]);
    });

    it("should have zero users count", () => {
      expect(usersStore.usersCount).toBe(0);
    });

    it("should have empty current filter", () => {
      expect(usersStore.currentFilter).toBe("");
    });
  });

  describe("setFilter", () => {
    it("should set filter value", () => {
      usersStore.setFilter("status:confirmed");
      expect(usersStore.currentFilter).toBe("status:confirmed");
    });

    it("should set empty string when filter is empty", () => {
      usersStore.setFilter("");
      expect(usersStore.currentFilter).toBe("");
    });
  });

  describe("fetchUsersList", () => {
    const baseUrl = "http://localhost:3000/admin/api/users";

    it("should fetch users list successfully with default pagination", async () => {
      const usersList = [mockUserBase];

      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" })).reply(200, usersList, { "x-total-count": "1" });

      await expect(usersStore.fetchUsersList()).resolves.not.toThrow();

      expect(usersStore.users).toEqual(usersList);
      expect(usersStore.usersCount).toBe(1);
    });

    it("should fetch users list successfully with custom pagination", async () => {
      const usersList = [mockUserBase, { ...mockUserBase, id: "user-id-456" }];

      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "2", per_page: "20" })).reply(200, usersList, { "x-total-count": "2" });

      await expect(usersStore.fetchUsersList({ page: 2, perPage: 20 })).resolves.not.toThrow();

      expect(usersStore.users).toEqual(usersList);
      expect(usersStore.usersCount).toBe(2);
    });

    it("should fetch users list with filter successfully", async () => {
      const usersList = [mockUserBase];
      const filter = "test";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter, page: "1", per_page: "10" })).reply(200, usersList, { "x-total-count": "1" });

      await expect(usersStore.fetchUsersList({ filter })).resolves.not.toThrow();

      expect(usersStore.users).toEqual(usersList);
      expect(usersStore.usersCount).toBe(1);
    });

    it("should use current filter when not provided in parameters", async () => {
      usersStore.setFilter("old_filter");

      const usersList = [mockUserBase];

      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter: "old_filter", page: "1", per_page: "10" }))
        .reply(200, usersList, { "x-total-count": "1" });

      await expect(usersStore.fetchUsersList()).resolves.not.toThrow();

      expect(usersStore.users).toEqual(usersList);
      expect(usersStore.usersCount).toBe(1);
    });

    it("should fetch empty users list successfully", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" })).reply(200, [], { "x-total-count": "0" });

      await expect(usersStore.fetchUsersList()).resolves.not.toThrow();

      expect(usersStore.users).toEqual([]);
      expect(usersStore.usersCount).toBe(0);
    });

    it("should throw on server error when fetching users list", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" })).reply(500);

      await expect(usersStore.fetchUsersList()).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when fetching users list", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" })).networkError();

      await expect(usersStore.fetchUsersList()).rejects.toThrow("Network Error");
    });
  });

  describe("exportUsersToCsv", () => {
    const baseUrl = "http://localhost:3000/admin/api/export/users";
    const csvData = "id,name,email,username\nuser-id-123,Admin User,admin@example.com,admin";

    it("should export users to CSV successfully and return data", async () => {
      const filter = "";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter })).reply(200, csvData);

      const result = await usersStore.exportUsersToCsv(filter);

      expect(result).toBe(csvData);
    });

    it("should export users with filter to CSV successfully", async () => {
      const filter = "admin:true";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter })).reply(200, csvData);

      const result = await usersStore.exportUsersToCsv(filter);

      expect(result).toBe(csvData);
    });

    it("should throw on server error when exporting users", async () => {
      const filter = "";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter })).reply(500);

      await expect(usersStore.exportUsersToCsv(filter)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when exporting users", async () => {
      const filter = "";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter })).networkError();

      await expect(usersStore.exportUsersToCsv(filter)).rejects.toThrow("Network Error");
    });
  });

  describe("addUser", () => {
    const baseUrl = "http://localhost:3000/admin/api/users";
    const { status: _status, ...userData } = mockUserFormData;
    it("should add user successfully", async () => {
      mockAdminApi.onPost(baseUrl, userData).reply(201);

      await expect(usersStore.addUser(mockUserFormData)).resolves.not.toThrow();
    });

    it("should throw on forbidden error when adding user", async () => {
      mockAdminApi.onPost(baseUrl, userData).reply(403, { message: "Forbidden" });

      await expect(usersStore.addUser(mockUserFormData)).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should throw on network error when adding user", async () => {
      mockAdminApi.onPost(baseUrl, userData).networkError();

      await expect(usersStore.addUser(mockUserFormData)).rejects.toThrow("Network Error");
    });
  });

  describe("fetchUserById", () => {
    const userId = "user-id-123";
    const baseGetUserUrl = `http://localhost:3000/admin/api/users/${userId}`;

    it("should fetch user by id successfully and return data", async () => {
      mockAdminApi.onGet(baseGetUserUrl).reply(200, mockUserBase);

      const result = await usersStore.fetchUserById(userId);

      expect(result).toEqual(mockUserBase);
    });

    it("should throw on not found error when fetching user by id", async () => {
      mockAdminApi.onGet(baseGetUserUrl).reply(404, { message: "User not found" });

      await expect(usersStore.fetchUserById(userId)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when fetching user by id", async () => {
      mockAdminApi.onGet(baseGetUserUrl).networkError();

      await expect(usersStore.fetchUserById(userId)).rejects.toThrow("Network Error");
    });
  });

  describe("updateUser", () => {
    const userId = "user-id-123";
    const updateData = { ...mockUserFormData, id: userId };
    const baseUrl = `http://localhost:3000/admin/api/users/${userId}`;

    it("should update user successfully", async () => {
      mockAdminApi.onPut(baseUrl, mockUserFormData).reply(200);

      await expect(usersStore.updateUser(updateData)).resolves.not.toThrow();
    });

    it("should throw on not found error when updating user", async () => {
      mockAdminApi.onPut(baseUrl, mockUserFormData).reply(404, { message: "User not found" });

      await expect(usersStore.updateUser(updateData)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when updating user", async () => {
      mockAdminApi.onPut(baseUrl, mockUserFormData).networkError();

      await expect(usersStore.updateUser(updateData)).rejects.toThrow("Network Error");
    });
  });

  describe("deleteUser", () => {
    const userId = "user-id-123";
    const baseUrl = `http://localhost:3000/admin/api/users/${userId}`;

    it("should delete user successfully", async () => {
      mockAdminApi.onDelete(baseUrl).reply(200);

      await expect(usersStore.deleteUser(userId)).resolves.not.toThrow();
    });

    it("should throw on not found error when deleting user", async () => {
      mockAdminApi.onDelete(baseUrl).reply(404, { message: "User not found" });

      await expect(usersStore.deleteUser(userId)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when deleting user", async () => {
      mockAdminApi.onDelete(baseUrl).networkError();

      await expect(usersStore.deleteUser(userId)).rejects.toThrow("Network Error");
    });
  });

  describe("resetUserPassword", () => {
    const userId = "user-id-123";
    const baseUrl = `http://localhost:3000/admin/api/users/${userId}/password/reset`;
    const newPassword = "new-temp-password-123";

    it("should reset user password successfully and return new password", async () => {
      mockAdminApi.onPatch(baseUrl).reply(200, newPassword);

      const result = await usersStore.resetUserPassword(userId);

      expect(result).toBe(newPassword);
    });

    it("should throw on server error when resetting user password", async () => {
      mockAdminApi.onPatch(baseUrl).reply(500);

      await expect(usersStore.resetUserPassword(userId)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when resetting user password", async () => {
      mockAdminApi.onPatch(baseUrl).networkError();

      await expect(usersStore.resetUserPassword(userId)).rejects.toThrow("Network Error");
    });
  });
});
