import { Module } from "vuex";
import { State } from "./../index";
import * as apiTags from "../api/tags";
import * as apiDevice from "../api/devices";

export interface TagsState {
  tags: Array<string>;
  numberTags: number;
  selected: Array<string>;
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
      state.selected = [];
    },
  },

  actions: {
    post: async (context, data) => {
      try {
        await apiDevice.postTag(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    setSelected: async (context, data) => {
      context.commit("setSelected", data);
    },

    fetch: async (context) => {
      try {
        const res = await apiTags.getTags();

        context.commit("setTags", res);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    edit: async (context, data) => {
      try {
        await apiTags.updateTag(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    setTags: (context, data) => {
      context.commit("setTags", data);
    },

    remove: async (context, name) => {
      try {
        await apiTags.removeTag(name);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    clearSelectedTags: (context) => {
      context.commit("clearSelected");
    },
  },
};
