import { Module } from "vuex";
import { AxiosResponse } from "axios";
import getStats from "../api/stats";
import { IStats } from "../../interfaces/IStats";

export interface StatsState {
  stats: IStats;
}

export function createStatsModule() {
  const stats: Module<StatsState, any> = {
    namespaced: true,
    state: {
      stats: {} as IStats,
    },

    getters: {
      stats: (state) => state.stats,
    },

    mutations: {
      setStats: (state, res: AxiosResponse) => {
        state.stats = res.data;
      },

      clearListState: (state) => {
        state.stats = {} as IStats;
      },
    },

    actions: {
      async get({ commit }) {
        try {
          const res = await getStats();
          commit("setStats", res);
          return res;
        } catch (error) {
          console.error(error);
          throw error;
        }
      },

      async clear({ commit }) {
        commit("clearListState");
      },
    },
  };
  return stats;
}
