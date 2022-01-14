import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    layout: 'appLayout',
    statusDarkMode: localStorage.getItem('statusDarkMode') === 'true' || localStorage.getItem('statusDarkMode') === null,
    statusNavigationDrawer: true,
  },

  getters: {
    getLayout: (state) => state.layout,
    getStatusDarkMode: (state) => state.statusDarkMode,
    getStatusNavigationDrawer: (state) => state.statusNavigationDrawer,
  },

  mutations: {
    setLayout: (state, layout) => {
      Vue.set(state, 'layout', layout);
    },

    setStatusDarkMode: (state, status) => {
      Vue.set(state, 'statusDarkMode', status);
    },

    setStatusNavigationDrawer: (state, status) => {
      Vue.set(state, 'statusNavigationDrawer', status);
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

    setStatusNavigationDrawer(context, status) {
      context.commit('setStatusNavigationDrawer', status);
    },
  },
};
