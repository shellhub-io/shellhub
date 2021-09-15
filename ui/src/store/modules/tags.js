import Vue from 'vue';
import * as apiDevice from '@/store/api/devices';

export default {
  namespaced: true,

  state: {
    tags: [],
    numberTags: 0,
  },

  getters: {
    list: (state) => state.tags,
    getNumberTags: (state) => state.numberTags,
  },

  mutations: {
    setTags: (state, res) => {
      Vue.set(state, 'tags', res.data);
      Vue.set(state, 'numberTags', parseInt(res.headers['x-total-count'], 10));
    },
  },

  actions: {
    post: async (context, data) => {
      await apiDevice.postTag(data);
    },

    fetch: async (context) => {
      const res = await apiDevice.getTags();

      context.commit('setTags', res);
    },

    edit: async (context, data) => {
      await apiDevice.updateTag(data);
    },

    remove: async (context, name) => {
      await apiDevice.removeTag(name);
    },
  },
};
