import Vue from 'vue';
import * as apiDevice from '@/store/api/devices';

export default {
  namespaced: true,

  state: {
    tags: [],
    numberTags: 0,
    selected: [],
  },

  getters: {
    list: (state) => state.tags,
    getNumberTags: (state) => state.numberTags,
    selected: (state) => state.selected,
  },

  mutations: {
    setTags: (state, res) => {
      Vue.set(state, 'tags', res.data);
      Vue.set(state, 'numberTags', parseInt(res.headers['x-total-count'], 10));
    },

    setSelected: (state, data) => {
      Vue.set(state, 'selected', data);
    },
  },

  actions: {
    post: async (context, data) => {
      await apiDevice.postTag(data);
    },

    setSelected: async (context, data) => {
      context.commit('setSelected', data);
    },

    fetch: async (context) => {
      const res = await apiDevice.getTags();

      context.commit('setTags', res);
    },

    edit: async (context, data) => {
      await apiDevice.updateTag(data);
    },

    setTags: (context, data) => {
      context.commit('setTags', data);
    },

    remove: async (context, name) => {
      await apiDevice.removeTag(name);
    },

    clearSelectedTags: (context) => {
      context.commit('setSelected', []);
    },
  },
};
