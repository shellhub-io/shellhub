import { Module } from "vuex";
import { State } from "./../index";

export interface MobileState {
  isMobile: boolean;
}

export const mobile: Module<MobileState, State> = {
  namespaced: true,
  state: {
    isMobile: false,
  },

  getters: {
    isMobile: (state) => state.isMobile,
  },

  mutations: {
    setIsMobileStatus: (state, status) => {
      state.isMobile = status
    },
  },

  actions: {
    setIsMobileStatus(context, status) {
      context.commit('setIsMobileStatus', status);
    },
  },
};
