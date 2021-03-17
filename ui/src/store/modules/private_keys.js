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

    editPrivateKey: (state, data) => {
      const { index, ...pk } = data;
      state.privateKeys.splice(index, 1, pk);
      localStorage.setItem('privateKeys', JSON.stringify(state.privateKeys));
    },

    removePrivateKey: (state, data) => {
      state.privateKeys.splice(state.privateKeys.findIndex((d) => d.data === data), 1);
      localStorage.setItem('privateKeys', JSON.stringify(state.privateKeys));
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

        privateKeys.forEach((pk) => {
          if (pk.data === privateKey.data && pk.name === privateKey.name) {
            throw new Error('both');
          }
          if (pk.data === privateKey.data) {
            throw new Error('private_key');
          }
          if (pk.name === privateKey.name) {
            throw new Error('name');
          }
        });
        privateKeys.push(privateKey);
        localStorage.setItem('privateKeys', JSON.stringify(privateKeys));
        context.commit('setPrivateKey', privateKey);

        return Promise.resolve();
      } catch (e) {
        return Promise.reject(e);
      }
    },

    edit: async (context, privateKey) => {
      try {
        let index;
        context.state.privateKeys.forEach((pk, i) => {
          if (pk.data === privateKey.data) {
            index = i;
          }
          if (pk.name === privateKey.name) {
            throw new Error('name');
          }
        });
        context.commit('editPrivateKey', { ...privateKey, ...{ index } });
        return Promise.resolve();
      } catch (e) {
        return Promise.reject(e);
      }
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
