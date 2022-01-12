import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    layout: 'appLayout',
    statusDarkMode: localStorage.getItem('statusDarkMode') === 'true' || localStorage.getItem('statusDarkMode') === null,
  },

  getters: {
    getLayout: (state) => state.layout,
    getStatusDarkMode: (state) => state.statusDarkMode,
  },

  mutations: {
    setLayout: (state, layout) => {
      Vue.set(state, 'layout', layout);
    },

    setStatusDarkMode: (state, status) => {
      Vue.set(state, 'statusDarkMode', status);
    },
  },

  actions: {
    setLayout(context, layout) {
      context.commit('setLayout', layout);
    },

    setStatusDarkMode(context, status) {
      context.commit('setStatusDarkMode', status);
      localStorage.setItem('statusDarkMode', status);
    },
  },
};
