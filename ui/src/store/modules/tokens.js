import Vue from 'vue';
import * as apiToken from '@/store/api/tokens';

export default {
  namespaced: true,

  state: {
    tokens: [],
    token: {},
    numberTokens: 0,
  },

  getters: {
    list: (state) => state.tokens,
    get: (state) => state.token,
    getNumberPublicKeys: (state) => state.numberTokens,
  },

  mutations: {
    setTokens: (state, res) => {
      Vue.set(state, 'tokens', res.data);
      Vue.set(state, 'numberTokens', parseInt(res.headers['x-total-count'], 10));
    },

    setToken: (state, res) => {
      Vue.set(state, 'token', res.data);
    },
  },

  actions: {
    post: async () => {
      await apiToken.postToken();
    },

    fetch: async (context) => {
      const res = await apiToken.fetchTokens();
      context.commit('setTokens', res);
    },

    get: async (context, id) => {
      const res = await apiToken.getToken(id);
      context.commit('setToken', res);
    },

    put: async (context, data) => {
      await apiToken.putToken(data);
    },

    remove: async (context, id) => {
      await apiToken.removeToken(id);
    },
  },
};
