import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    status: false,
  },

  getters: {
    getStatus: (state) => state.status,
  },

  mutations: {
    setStatus: (state, status) => {
      Vue.set(state, 'status', status);
    },
  },

  actions: {
    setStatus: async (context, status) => {
      context.commit('setStatus', status);
    },
  },
};
