export default {
  namespaced: true,

  state: {
    terminal: '',
    addDevice: false,
    screenWelcome: false,
  },

  getters: {
    terminal: (state) => {
      return state.terminal;
    },

    addDevice: (state) => {
      return state.addDevice;
    },

    statusScreenWelcome: (state) => {
      return state.screenWelcome;
    }
  },

  mutations: {
    setTerminal: (state, data) => {
      state.terminal = data;
    },

    setAddDevice: (state, data) => {
      state.addDevice = data;
    },

    setUserWelcome: (state, data) => {
      state.screenWelcome = data;
    },
  },

  actions: {
    toggleTerminal: (context, value) => {
      context.commit('setTerminal', value);
    },

    showAddDevice: (context, value) => {
      context.commit('setAddDevice', value);
    },

    showUserWelcome: (context, value) => {
      context.commit('setUserWelcome', value);
    }
  }
};