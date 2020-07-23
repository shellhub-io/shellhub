import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    terminal: '',
    addDevice: false,
  },

  getters: {
    terminal: (state) => state.terminal,
    addDevice: (state) => state.addDevice,
  },

  mutations: {
    setTerminal: (state, data) => {
      Vue.set(state, 'terminal', data);
    },

    setAddDevice: (state, data) => {
      Vue.set(state, 'addDevice', data);
    },
  },

  actions: {
    toggleTerminal: (context, value) => {
      context.commit('setTerminal', value);
    },

    showAddDevice: (context, value) => {
      context.commit('setAddDevice', value);
    },
  },
};
