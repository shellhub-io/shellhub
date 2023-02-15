import { Module } from "vuex";
import { State } from "./../index";

export interface SpinnerState {
  status: boolean;
}

export const spinner: Module<SpinnerState, State> = {
  namespaced: true,
  state: {
    status: false,
  },
  getters: {
    status: (state: SpinnerState) => state.status,
  },
  mutations: {
    setStatus: (state, status) => {
      state.status = status;
    },
  },

  actions: {
    setStatus: async ({ commit }, status) => {
      commit("setStatus", status);
    },
  },
};
