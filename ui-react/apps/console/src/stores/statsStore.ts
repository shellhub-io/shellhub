import { create } from "zustand";
import { getStats } from "@/api/stats";
import { Stats } from "@/types/stats";

interface StatsState {
  stats: Stats | null;
  loading: boolean;
  error: boolean;
  fetch: () => Promise<void>;
}

export const useStatsStore = create<StatsState>((set) => ({
  stats: null,
  loading: true,
  error: false,

  fetch: async () => {
    set({ loading: true, error: false });
    try {
      const stats = await getStats();
      set({ stats, loading: false });
    } catch {
      set({ loading: false, error: true });
    }
  },
}));
