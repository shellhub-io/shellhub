import Vue from 'vue';
import * as apiTags from '@/store/api/tags';
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
      const res = await apiTags.getTags();

      context.commit('setTags', res);
    },

    edit: async (context, data) => {
      await apiTags.updateTag(data);
    },

    setTags: (context, data) => {
      context.commit('setTags', data);
    },

    remove: async (context, name) => {
      await apiTags.removeTag(name);
    },

    clearSelectedTags: (context) => {
      context.commit('setSelected', []);
    },
  },
};
