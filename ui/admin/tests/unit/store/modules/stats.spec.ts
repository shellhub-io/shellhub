import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import useStatsStore from "@admin/store/modules/stats";
import { IAdminStats } from "@admin/interfaces/IStats";

const mockStats: IAdminStats = {
  registered_devices: 10,
  online_devices: 5,
  active_sessions: 3,
  pending_devices: 2,
  rejected_devices: 1,
  registered_users: 15,
};

describe("Admin Stats Store", () => {
  let statsStore: ReturnType<typeof useStatsStore>;
  let mockAdminApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    statsStore = useStatsStore();
    mockAdminApi = new MockAdapter(adminApi.getAxios());
  });

  afterEach(() => { mockAdminApi.reset(); });

  describe("initial state", () => {
    it("should have empty stats initially", () => {
      expect(statsStore.stats).toEqual({} as IAdminStats);
    });
  });

  describe("getStats", () => {
    const baseUrl = "http://localhost:3000/admin/api/stats";

    it("should fetch stats successfully and return data", async () => {
      mockAdminApi.onGet(baseUrl).reply(200, mockStats);

      await statsStore.getStats();

      expect(statsStore.stats).toEqual(mockStats);
    });

    it("should throw on not found error when fetching stats", async () => {
      mockAdminApi.onGet(baseUrl).reply(404, { message: "Stats not found" });

      await expect(statsStore.getStats()).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when fetching stats", async () => {
      mockAdminApi.onGet(baseUrl).networkError();

      await expect(statsStore.getStats()).rejects.toThrow("Network Error");
    });
  });
});
