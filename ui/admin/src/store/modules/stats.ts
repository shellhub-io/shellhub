import { defineStore } from "pinia";
import getStats from "../api/stats";
import { IStats } from "../../interfaces/IStats";

export const useStatsStore = defineStore("stats", {
  state: () => ({
    stats: {} as IStats,
  }),

  getters: {
    getStats: (state) => state.stats,
  },

  actions: {
    async get() {
      const res = await getStats();
      this.stats = res.data as IStats;
      return res;
    },

    clearListState() {
      this.stats = {} as IStats;
    },
  },
});

export default useStatsStore;
