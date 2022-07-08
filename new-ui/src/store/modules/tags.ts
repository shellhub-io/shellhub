import { Module } from "vuex";
import { State } from "./../index";
import * as apiTags from "../api/tags";
import * as apiDevice from "../api/devices";

export interface TagsState {
  tags: Array<any>;
  numberTags: number;
  selected: Array<any>;
}

export const tags: Module<TagsState, State> = {
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
      state.tags = res.data;
      state.numberTags = parseInt(res.headers["x-total-count"], 10);
    },

    setSelected: (state, data) => {
      if (state.selected.includes(data)) {
        state.selected.splice(state.selected.indexOf(data), 1);
      } else {
        state.selected = [...state.selected, data];
      }
    },
    clearSelected: (state) => {
      state.selected = []
    }
  },

  actions: {
    post: async (context, data) => {
      await apiDevice.postTag(data);
    },

    setSelected: async (context, data) => {
      context.commit("setSelected", data);
    },

    fetch: async (context) => {
      const res = await apiTags.getTags();

      context.commit("setTags", res);
    },

    edit: async (context, data) => {
      await apiTags.updateTag(data);
    },

    setTags: (context, data) => {
      context.commit("setTags", data);
    },

    remove: async (context, name) => {
      await apiTags.removeTag(name);
    },

    clearSelectedTags: (context) => {
      context.commit("clearSelected");
    },
  },
};
