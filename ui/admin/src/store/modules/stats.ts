import { defineStore } from "pinia";
import { IAdminStats } from "@admin/interfaces/IStats";
import getAdminStats from "../api/stats";

const useStatsStore = defineStore("stats", () => {
  const getStats = async () => {
    const { data } = await getAdminStats();
    return data as IAdminStats;
  };

  return { getStats };
});

export default useStatsStore;
