import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    isMobile: false,
  },

  getters: {
    isMobile: (state) => state.isMobile,
  },

  mutations: {
    setIsMobileStatus: (state, status) => {
      Vue.set(state, 'isMobile', status);
    },
  },

  actions: {
    setIsMobileStatus(context, status) {
      context.commit('setIsMobileStatus', status);
    },
  },
};
