import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useStatsStore from "@admin/store/modules/stats";

describe("Stats Pinia Store", () => {
  let statsStore: ReturnType<typeof useStatsStore>;

  const stats = {
    registered_devices: 2,
    online_devices: 1,
    active_sessions: 1,
    pending_devices: 1,
    registered_users: 10,
    rejected_devices: 0,
  };

  beforeEach(() => {
    setActivePinia(createPinia());
    statsStore = useStatsStore();
  });

  it("returns default stats state", () => {
    expect(statsStore.getStats).toEqual({});
  });

  it("sets stats state", () => {
    statsStore.stats = stats;
    expect(statsStore.getStats).toEqual(stats);
  });

  it("clears stats state", () => {
    statsStore.stats = stats;
    statsStore.clearListState();
    expect(statsStore.getStats).toEqual({});
  });
});
