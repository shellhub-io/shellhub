import { defineStore } from "pinia";
import { IAdminStats } from "@admin/interfaces/IStats";
import getStats from "../api/stats";

export const useStatsStore = defineStore("stats", {
  state: () => ({
    stats: {} as IAdminStats,
  }),

  getters: {
    getStats: (state) => state.stats,
  },

  actions: {
    async get() {
      const res = await getStats();
      this.stats = res.data as IAdminStats;
      return res;
    },

    clearListState() {
      this.stats = {} as IAdminStats;
    },
  },
});

export default useStatsStore;
