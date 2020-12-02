import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    privateKeys: [],
    numberPrivateKeys: 0,
  },

  getters: {
    list: (state) => state.privateKeys,
    getNumberPrivateKeys: (state) => state.numberPrivateKeys,
  },

  mutations: {
    fetchPrivateKey: (state, privateKey) => {
      Vue.set(state, 'privateKeys', privateKey);
      Vue.set(state, 'numberPrivateKeys', privateKey.length);
    },

    setPrivateKey: (state, privateKey) => {
      let { numberPrivateKeys } = state;

      state.privateKeys.push(privateKey);
      Vue.set(state, 'numberPrivateKeys', numberPrivateKeys += 1);
    },

    editPrivateKey: (state, privateKey) => {
      Vue.set(state, 'privateKeys', state.privateKeys.map((i) => (i.data === privateKey.data ? { ...i, name: privateKey.name } : i)));
    },

    removePrivateKey: (state, data) => {
      state.privateKeys.splice(state.privateKeys.findIndex((d) => d.data === data), 1);
    },

    clearListPrivateKeys: (state) => {
      Vue.set(state, 'privateKeys', []);
      Vue.set(state, 'numberPrivateKeys', 0);
    },
  },

  actions: {
    fetch: async (context) => {
      const privateKeys = JSON.parse(localStorage.getItem('privateKeys'));
      if (privateKeys !== null) { context.commit('fetchPrivateKey', privateKeys); }
    },

    set: async (context, privateKey) => {
      try {
        const privateKeys = JSON.parse(localStorage.getItem('privateKeys')) || [];

        if (privateKeys.find((element) => element.data === privateKey.data) !== undefined) {
          throw new Error();
        } else {
          privateKeys.push(privateKey);
          localStorage.setItem('privateKeys', JSON.stringify(privateKeys));
          context.commit('setPrivateKey', privateKey);

          return true;
        }
      } catch (e) {
        return Promise.reject(e);
      }
    },

    edit: async (context, privateKey) => {
      context.commit('editPrivateKey', privateKey);
    },

    remove: async (context, data) => {
      const privateKeys = JSON.parse(localStorage.getItem('privateKeys')) || [];

      if (privateKeys !== null) {
        privateKeys.splice(privateKeys.findIndex((d) => d.data === data), 1);
      }

      localStorage.setItem('privateKeys', JSON.stringify(privateKeys));
      context.commit('removePrivateKey', data);
    },
  },
};
