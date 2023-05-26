import { Module } from "vuex";
import { State } from "..";

export interface LayoutState {
  layout: string;
  statusDarkMode: string;
}

export const layout: Module<LayoutState, State> = {
  namespaced: true,
  state: {
    layout: "appLayout",
    statusDarkMode: localStorage.getItem("statusDarkMode") || "dark",
  },
  getters: {
    getLayout: (state) => state.layout,
    getStatusDarkMode: (state) => state.statusDarkMode,
  },
  mutations: {
    setLayout: (state, layout) => {
      state.layout = layout;
    },
    setStatusDarkMode: (state, status) => {
      state.statusDarkMode = status;
    },
  },

  actions: {
    setLayout({ commit }, layout) {
      commit("setLayout", layout);
    },

    setStatusDarkMode({ commit }, status: boolean) {
      const statusDarkMode = status ? "dark" : "light";
      commit("setStatusDarkMode", statusDarkMode);
      localStorage.setItem("statusDarkMode", statusDarkMode);
    },
  },
};
