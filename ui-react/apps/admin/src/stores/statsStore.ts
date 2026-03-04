import { create } from "zustand";
import { getStats } from "@/api/stats";
import { Stats } from "@/types/stats";

interface StatsState {
  stats: Stats | null;
  fetch: () => Promise<void>;
}

export const useStatsStore = create<StatsState>((set) => ({
  stats: null,

  fetch: async () => {
    try {
      const stats = await getStats();
      set({ stats });
    } catch {
      // Stats unavailable — fail silently
    }
  },
}));
