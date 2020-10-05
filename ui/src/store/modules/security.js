import Vue from 'vue';
import * as apiUser from '@/store/api/users';

export default {
  namespaced: true,

  state: {
    sessionRecord: true,
  },

  getters: {
    get: (state) => state.sessionRecord,
  },

  mutations: {
    setSecurity: (state, res) => {
      Vue.set(state, 'sessionRecord', res);
    },
  },

  actions: {
    async set(context, status) {
      await apiUser.putSecurity(status);
      context.commit('setSecurity', status);
    },

    async get(context) {
      const res = await apiUser.getSecurity();
      context.commit('setSecurity', res.data);
    },
  },
};
