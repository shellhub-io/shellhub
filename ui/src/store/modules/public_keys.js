import Vue from 'vue';
import * as apiPublicKey from '@/store/api/public_keys';

export default {
  namespaced: true,

  state: {
    publicKeys: [],
    publicKey: {},
    numberPublicKeys: 0,
    page: 0,
    perPage: 10,
  },

  getters: {
    list: (state) => state.publicKeys,
    get: (state) => state.publicKey,
    getNumberPublicKeys: (state) => state.numberPublicKeys,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
  },

  mutations: {
    setPublicKeys: (state, res) => {
      Vue.set(state, 'publicKeys', res.data);
      Vue.set(state, 'numberPublicKeys', parseInt(res.headers['x-total-count'], 10));
    },

    setPublicKey: (state, res) => {
      Vue.set(state, 'publicKey', res.data);
    },

    setPagePerpage: (state, data) => {
      Vue.set(state, 'page', data.page);
      Vue.set(state, 'perPage', data.perPage);
    },

    resetPagePerpage: (state) => {
      Vue.set(state, 'page', 0);
      Vue.set(state, 'perPage', 10);
    },

    removePublicKey: (state, fingerprint) => {
      state.publicKeys.splice(state.publicKeys.findIndex((d) => d.fingerprint === fingerprint), 1);
    },

    clearListPublicKeys: (state) => {
      Vue.set(state, 'publicKeys', []);
      Vue.set(state, 'numberPublicKeys', 0);
    },

    clearObjectPublicKey: (state) => {
      Vue.set(state, 'publicKey', {});
    },
  },

  actions: {
    post: async (context, data) => {
      await apiPublicKey.postPublicKey(data);
    },

    fetch: async (context, data) => {
      try {
        const res = await apiPublicKey.fetchPublicKeys(data.perPage, data.page);
        context.commit('setPublicKeys', res);
        context.commit('setPagePerpage', data);
      } catch (error) {
        context.commit('clearListPublicKeys');
        throw error;
      }
    },

    refresh: async (context) => {
      try {
        const res = await apiPublicKey.fetchPublicKeys(
          context.state.perPage,
          context.state.page,
        );
        context.commit('setPublicKeys', res);
      } catch (error) {
        context.commit('clearListPublicKeys');
        throw error;
      }
    },

    get: async (context, id) => {
      try {
        const res = await apiPublicKey.getPublicKey(id);
        context.commit('setPublicKey', res);
      } catch (error) {
        context.commit('clearObjectPublicKey');
        throw error;
      }
    },

    put: async (context, data) => {
      await apiPublicKey.putPublicKey(data);
    },

    resetPagePerpage: async (context) => {
      context.commit('resetPagePerpage');
    },

    remove: async (context, fingerprint) => {
      await apiPublicKey.removePublicKey(fingerprint);
    },
  },
};
