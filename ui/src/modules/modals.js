export default {
  namespaced: true,

  state: {
    terminal: '',
    addDevice: false,
  },

  getters: {
    terminal: (state) => {
      return state.terminal;
    },

    addDevice: (state) => {
      return state.addDevice;
    }
  },

  mutations: {
    setTerminal: (state, data) => {
      state.terminal = data;
    },

    setAddDevice: (state, data) => {
      state.addDevice = data;
    }
  },

  actions: {
    toggleTerminal: (context, value) => {
      context.commit('setTerminal', value);
    },

    showAddDevice: (context, value) => {
      context.commit('setAddDevice', value);
    }
  }
};
