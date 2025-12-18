import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { devicesApi } from "@/api/http";
import { IStats } from "@/interfaces/IStats";
import useStatsStore from "@/store/modules/stats";

const mockStatsBase: IStats = {
  registered_devices: 10,
  online_devices: 5,
  active_sessions: 3,
  pending_devices: 2,
  rejected_devices: 1,
};

describe("Stats Store", () => {
  let statsStore: ReturnType<typeof useStatsStore>;
  let mockDevicesApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    statsStore = useStatsStore();
    mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  });

  afterEach(() => { mockDevicesApi.reset(); });

  describe("Initial State", () => {
    it("should have empty stats object", () => {
      expect(statsStore.stats).toEqual({});
    });
  });

  describe("fetchStats", () => {
    const statsUrl = "http://localhost:3000/api/stats";

    it("should fetch stats successfully and update state", async () => {
      mockDevicesApi
        .onGet(statsUrl)
        .reply(200, mockStatsBase);

      await expect(statsStore.fetchStats()).resolves.not.toThrow();

      expect(statsStore.stats).toEqual(mockStatsBase);
    });

    it("should fetch stats with all zero values", async () => {
      const zeroStats: IStats = {
        registered_devices: 0,
        online_devices: 0,
        active_sessions: 0,
        pending_devices: 0,
        rejected_devices: 0,
      };

      mockDevicesApi
        .onGet(statsUrl)
        .reply(200, zeroStats);

      await expect(statsStore.fetchStats()).resolves.not.toThrow();

      expect(statsStore.stats).toEqual(zeroStats);
    });

    it("should handle not found error when fetching stats", async () => {
      mockDevicesApi
        .onGet(statsUrl)
        .reply(404, { message: "Stats not found" });

      await expect(statsStore.fetchStats()).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when fetching stats", async () => {
      mockDevicesApi
        .onGet(statsUrl)
        .reply(500);

      await expect(statsStore.fetchStats()).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when fetching stats", async () => {
      mockDevicesApi
        .onGet(statsUrl)
        .networkError();

      await expect(statsStore.fetchStats()).rejects.toThrow("Network Error");
    });
  });
});
