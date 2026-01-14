import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminStats } from "@admin/interfaces/IStats";
import getAdminStats from "../api/stats";

const useStatsStore = defineStore("adminStats", () => {
  const stats = ref<IAdminStats>({} as IAdminStats);
  const getStats = async () => {
    const { data } = await getAdminStats();
    stats.value = data as IAdminStats;
  };

  return { stats, getStats };
});

export default useStatsStore;
