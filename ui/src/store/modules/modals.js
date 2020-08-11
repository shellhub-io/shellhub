import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    terminal: '',
    addDevice: false,
    snackbarError: false,
  },

  getters: {
    terminal: (state) => state.terminal,
    addDevice: (state) => state.addDevice,
    snackbarError: (state) => state.snackbarError,
  },

  mutations: {
    setTerminal: (state, data) => {
      Vue.set(state, 'terminal', data);
    },

    setAddDevice: (state, data) => {
      Vue.set(state, 'addDevice', data);
    },

    setSnackbarError: (state, data) => {
      Vue.set(state, 'snackbarError', data);
    },
  },

  actions: {
    toggleTerminal: (context, value) => {
      context.commit('setTerminal', value);
    },

    showAddDevice: (context, value) => {
      context.commit('setAddDevice', value);
    },

    showSnackbarError: (context, value) => {
      context.commit('setSnackbarError', value);
    },
  },
};
