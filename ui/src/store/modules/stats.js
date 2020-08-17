import Vue from 'vue';
import getStats from '@/store/api/stats';

export default {
  namespaced: true,

  state: {
    stats: [],
  },

  getters: {
    stats: (state) => state.stats,
  },

  mutations: {
    setStats: (state, res) => {
      Vue.set(state, 'stats', res.data);
    },

    clearListState: (state) => {
      Vue.set(state, 'stats', []);
    },
  },

  actions: {
    get: async (context) => {
      try {
        const res = await getStats();
        context.commit('setStats', res);
      } catch (error) {
        context.commit('clearListState');
        throw error;
      }
    },
  },
};
