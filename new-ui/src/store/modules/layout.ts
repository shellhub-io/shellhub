import { Module } from "vuex";
import { State } from "./../index";

export interface LayoutState {
  layout: string;
  statusDarkMode: string;
  statusNavigationDrawer: boolean;
}

export const layout: Module<LayoutState, State> = {
  namespaced: true,
  state: {
    layout: "appLayout",
    statusDarkMode: localStorage.getItem("statusDarkMode") || "dark",
    statusNavigationDrawer: true,
  },
  getters: {
    getLayout: (state) => state.layout,
    getStatusDarkMode: (state) => state.statusDarkMode,
    getStatusNavigationDrawer: (state) => state.statusNavigationDrawer,
  },
  mutations: {
    setLayout: (state, layout) => {
      state.layout = layout;
    },
    setStatusDarkMode: (state, status) => {
      state.statusDarkMode = status;
    },
    setStatusNavigationDrawer: (state, status) => {
      state.statusNavigationDrawer = status;
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

    setStatusNavigationDrawer(context, status) {
      context.commit("setStatusNavigationDrawer", status);
    },
  },
};
