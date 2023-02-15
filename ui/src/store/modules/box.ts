import { Module } from "vuex";
import { State } from "./../index";

export interface BoxState {
  status: boolean;
}

export const box: Module<BoxState, State> = {
  namespaced: true,
  state: {
    status: true,
  },

  getters: {
    getStatus: (state) => state.status,
  },

  mutations: {
    setStatus: (state, status) => {
      state.status = status;
    },
  },

  actions: {
    setStatus: async (context, status) => {
      context.commit("setStatus", status);
    },
  },
};
