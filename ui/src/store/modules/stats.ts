import { defineStore } from "pinia";
import { ref } from "vue";
import getStats from "../api/stats";
import { IStats } from "@/interfaces/IStats";

const useStatsStore = defineStore("stats", () => {
  const stats = ref({} as IStats);

  const fetchStats = async () => {
    const res = await getStats();
    stats.value = res.data as IStats;
  };

  return {
    stats,
    fetchStats,
  };
});

export default useStatsStore;
