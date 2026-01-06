import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import useAuthStore from "@admin/store/modules/auth";

describe("Admin Auth Store", () => {
  let authStore: ReturnType<typeof useAuthStore>;
  let mockAdminApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockAdminApi = new MockAdapter(adminApi.getAxios());
    localStorage.clear();
    authStore = useAuthStore();
  });

  afterEach(() => {
    mockAdminApi.reset();
    localStorage.clear();
  });

  describe("Initial State", () => {
    it("should have empty status", () => {
      expect(authStore.status).toBe("");
    });

    it("should have empty token", () => {
      expect(authStore.token).toBe("");
    });

    it("should have empty current user", () => {
      expect(authStore.currentUser).toBe("");
    });

    it("should have isAdmin as false", () => {
      expect(authStore.isAdmin).toBe(false);
    });

    it("should have isLoggedIn as false when no token", () => {
      expect(authStore.isLoggedIn).toBe(false);
    });

    it("should load token from localStorage if present", () => {
      localStorage.setItem("token", "stored-token");
      localStorage.setItem("user", "stored-user");
      localStorage.setItem("admin", "true");

      setActivePinia(createPinia());
      const newStore = useAuthStore();

      expect(newStore.token).toBe("stored-token");
      expect(newStore.currentUser).toBe("stored-user");
      expect(newStore.isAdmin).toBe(true);
      expect(newStore.isLoggedIn).toBe(true);
    });
  });

  describe("getLoginToken", () => {
    const baseUrl = "http://localhost:3000/admin/api/auth/token";

    it("should fetch token successfully and return it", async () => {
      const userId = "user-123";
      const mockResponse = {
        token: "generated-token-abc123",
      };

      mockAdminApi.onGet(`${baseUrl}/${userId}`).reply(200, mockResponse);

      const result = await authStore.getLoginToken(userId);

      expect(result).toBe("generated-token-abc123");
      expect(authStore.status).toBe("");
    });

    it("should set status to error on 500 and throw", async () => {
      const userId = "user-123";

      mockAdminApi.onGet(`${baseUrl}/${userId}`).reply(500);

      await expect(authStore.getLoginToken(userId)).rejects.toBeAxiosErrorWithStatus(500);
      expect(authStore.status).toBe("error");
    });

    it("should set status to error on network error and throw", async () => {
      const userId = "user-123";

      mockAdminApi.onGet(`${baseUrl}/${userId}`).networkError();

      await expect(authStore.getLoginToken(userId)).rejects.toThrow("Network Error");
      expect(authStore.status).toBe("error");
    });
  });

  describe("logout", () => {
    it("should clear all state and localStorage", () => {
      authStore.status = "success";
      authStore.token = "test-token";
      authStore.currentUser = "test-user";
      localStorage.setItem("token", "test-token");
      localStorage.setItem("user", "test-user");

      authStore.logout();

      expect(authStore.status).toBe("");
      expect(authStore.token).toBe("");
      expect(authStore.currentUser).toBe("");
      expect(authStore.isLoggedIn).toBe(false);
      expect(localStorage.getItem("token")).toBeNull();
      expect(localStorage.getItem("user")).toBeNull();
    });
  });
});
