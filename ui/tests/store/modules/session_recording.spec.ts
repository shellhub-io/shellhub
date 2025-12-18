import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { usersApi } from "@/api/http";
import useSessionRecordingStore from "@/store/modules/session_recording";

describe("Session Recording Store", () => {
  let sessionRecordingStore: ReturnType<typeof useSessionRecordingStore>;
  let mockUsersApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    sessionRecordingStore = useSessionRecordingStore();
    mockUsersApi = new MockAdapter(usersApi.getAxios());
  });

  afterEach(() => { mockUsersApi.reset(); });

  describe("Initial State", () => {
    it("should have session recording enabled by default", () => {
      expect(sessionRecordingStore.isEnabled).toBe(true);
    });
  });

  describe("getStatus", () => {
    const securityUrl = "http://localhost:3000/api/users/security";

    it("should get session recording status as enabled", async () => {
      mockUsersApi
        .onGet(securityUrl)
        .reply(200, true);

      await expect(sessionRecordingStore.getStatus()).resolves.not.toThrow();

      expect(sessionRecordingStore.isEnabled).toBe(true);
    });

    it("should get session recording status as disabled", async () => {
      mockUsersApi
        .onGet(securityUrl)
        .reply(200, false);

      await expect(sessionRecordingStore.getStatus()).resolves.not.toThrow();

      expect(sessionRecordingStore.isEnabled).toBe(false);
    });

    it("should handle not found error when getting status", async () => {
      mockUsersApi
        .onGet(securityUrl)
        .reply(404, { message: "User not found" });

      await expect(sessionRecordingStore.getStatus()).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when getting status", async () => {
      mockUsersApi
        .onGet(securityUrl)
        .reply(500);

      await expect(sessionRecordingStore.getStatus()).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when getting status", async () => {
      mockUsersApi
        .onGet(securityUrl)
        .networkError();

      await expect(sessionRecordingStore.getStatus()).rejects.toThrow("Network Error");
    });
  });

  describe("setStatus", () => {
    const updateSecurityUrl = (userId: string) => `http://localhost:3000/api/users/security/${userId}`;

    it("should set session recording status to enabled", async () => {
      const data = {
        id: "user-id-123",
        status: true,
      };

      mockUsersApi
        .onPut(updateSecurityUrl(data.id))
        .reply(200);

      await expect(sessionRecordingStore.setStatus(data)).resolves.not.toThrow();

      expect(sessionRecordingStore.isEnabled).toBe(true);
    });

    it("should set session recording status to disabled", async () => {
      const data = {
        id: "user-id-123",
        status: false,
      };

      mockUsersApi
        .onPut(updateSecurityUrl(data.id))
        .reply(200);

      await expect(sessionRecordingStore.setStatus(data)).resolves.not.toThrow();

      expect(sessionRecordingStore.isEnabled).toBe(false);
    });

    it("should handle not found error when setting status", async () => {
      const data = {
        id: "invalid-user-id",
        status: true,
      };

      mockUsersApi
        .onPut(updateSecurityUrl(data.id))
        .reply(404, { message: "User not found" });

      await expect(sessionRecordingStore.setStatus(data)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when setting status", async () => {
      const data = {
        id: "user-id-123",
        status: true,
      };

      mockUsersApi
        .onPut(updateSecurityUrl(data.id))
        .reply(500);

      await expect(sessionRecordingStore.setStatus(data)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when setting status", async () => {
      const data = {
        id: "user-id-123",
        status: true,
      };

      mockUsersApi
        .onPut(updateSecurityUrl(data.id))
        .networkError();

      await expect(sessionRecordingStore.setStatus(data)).rejects.toThrow("Network Error");
    });
  });
});
