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
    async set(context, data) {
      await apiUser.putSecurity(data);
      context.commit('setSecurity', data.status);
    },

    async get(context) {
      const res = await apiUser.getSecurity();
      context.commit('setSecurity', res.data);
    },
  },
};
