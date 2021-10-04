import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    layout: 'appLayout',
  },

  getters: {
    getLayout: (state) => state.layout,
  },

  mutations: {
    setLayout: (state, layout) => {
      Vue.set(state, 'layout', layout);
    },
  },

  actions: {
    setLayout(context, layout) {
      context.commit('setLayout', layout);
    },
  },
};
