import { Module } from "vuex";
import { State } from "./../index";
import getStats from "../api/stats";
import { IStats } from "../../interfaces/IStats";

export interface StatsState {
  stats: IStats | {};
}

export const stats: Module<StatsState, State> = {
  namespaced: true,

  state: {
    stats: {},
  },

  getters: {
    stats: (state) => state.stats,
  },

  mutations: {
    setStats: (state, res) => {
      state.stats = res.data;
    },

    clearListState: (state) => {
      state.stats = {};
    },
  },

  actions: {
    async get({ commit }) {
      const res = await getStats();
      commit("setStats", res);
      return res;
    },
  },
};
