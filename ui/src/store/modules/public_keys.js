import Vue from 'vue';
import {
  postPublicKey, fetchPublicKeys, getPublicKey, putPublicKey, removePublicKey,
} from '@/store/api/public_keys';

export default {
  namespaced: true,

  state: {
    publicKeys: [],
    publicKey: {},
    numberPublicKeys: 0,
  },

  getters: {
    list: (state) => state.publicKeys,
    get: (state) => state.publicKey,
    getNumberPublicKeys: (state) => state.numberPublicKeys,
  },

  mutations: {
    setPublicKeys: (state, res) => {
      Vue.set(state, 'publicKeys', res.data);
      Vue.set(state, 'numberPublicKeys', parseInt(res.headers['x-total-count'], 10));
    },

    setPublicKey: (state, res) => {
      Vue.set(state, 'publicKey', res.data);
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
      await postPublicKey(data);
    },

    fetch: async (context, data) => {
      try {
        const res = await fetchPublicKeys(data.perPage, data.page);
        context.commit('setPublicKeys', res);
      } catch (error) {
        context.commit('clearListPublicKeys');
        throw error;
      }
    },

    get: async (context, id) => {
      try {
        const res = await getPublicKey(id);
        context.commit('setPublicKey', res);
      } catch (error) {
        context.commit('clearObjectPublicKey');
        throw error;
      }
    },

    put: async (context, data) => {
      await putPublicKey(data);
    },

    remove: async (context, fingerprint) => {
      await removePublicKey(fingerprint);
      context.commit('removePublicKey', fingerprint);
    },
  },
};
