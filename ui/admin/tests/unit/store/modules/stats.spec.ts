import { describe, expect, it, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useStatsStore from "@admin/store/modules/stats";

const stats = {
  registered_devices: 2,
  online_devices: 1,
  active_sessions: 1,
  pending_devices: 1,
  registered_users: 10,
  rejected_devices: 0,
};

describe("Stats Pinia Store", () => {
  setActivePinia(createPinia());
  const statsStore = useStatsStore();

  it("returns default stats state", async () => {
    statsStore.getStats = vi.fn().mockResolvedValue(stats);
    const responseData = await statsStore.getStats();
    expect(responseData).toEqual(stats);
  });
});
