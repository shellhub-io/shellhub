import { Module } from "vuex";
import * as apiTags from "../api/tags";
import { State } from "..";
import { Tags } from "@/interfaces/ITags";

export interface TagsState {
  tags: Array<Tags>;
  page: number;
  perPage: number;
  filter: undefined | string;
  numberTags: number;
  selected: {
    device: Array<Tags>;
    container: Array<Tags>;
  };
}

export const tags: Module<TagsState, State> = {
  namespaced: true,
  state: {
    tags: [],
    numberTags: 0,
    page: 1,
    perPage: 10,
    filter: "",
    selected: {
      device: [],
      container: [],
    },
  },

  getters: {
    list: (state) => state.tags,
    getNumberTags: (state) => state.numberTags,
    getPerPage: (state) => state.perPage,
    getPage: (state) => state.page,
    getFilter: (state) => state.filter,
    selected: (state) => (variant: "device" | "container") => state.selected[variant],
  },

  mutations: {
    setTags: (state, res) => {
      state.tags = res.data;
      state.numberTags = parseInt(res.headers["x-total-count"], 10);
    },
    setPagePerPage: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
    },
    setFilter: (state, data) => {
      state.filter = data.filter;
    },
    clearListTags: (state) => {
      state.tags = [];
      state.numberTags = 0;
    },

    setSelected: (state, { variant, tag }) => {
      if (state.selected[variant].includes(tag)) {
        state.selected[variant] = state.selected[variant].filter((t) => t !== tag);
      } else {
        state.selected[variant] = [...state.selected[variant], tag];
      }
    },
    clearSelected: (state, variant: "device" | "container") => {
      state.selected[variant] = [];
    },
  },

  actions: {
    async createTag(context, { tenant, name }) {
      try {
        await apiTags.createTag(tenant, name);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async editTag(context, { tenant, currentName, newName }) {
      try {
        await apiTags.updateTag(tenant, currentName, newName);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async removeTag(context, { tenant, currentName }) {
      try {
        await apiTags.removeTag(tenant, currentName);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async fetch(context, { tenant, filter, page, perPage }) {
      try {
        const res = await apiTags.getTags(tenant, filter, page, perPage);
        context.commit("setTags", res);
        context.commit("setPagePerPage", { page, perPage });
        context.commit("setFilter", { filter });
      } catch (error) {
        context.commit("clearListTags");
        throw error;
      }
    },

    async search(context, { tenant, filter }) {
      try {
        const res = await apiTags.getTags(tenant, filter, context.state.page, context.state.perPage);
        context.commit("setTags", res);
        context.commit("setFilter", { filter });
      } catch (error) {
        context.commit("clearListTags");
        throw error;
      }
    },

    async autocomplete(context, { tenant, filter, page, perPage }) {
      try {
        const res = await apiTags.getTags(tenant, filter, page, perPage);
        context.commit("setTags", res);
        context.commit("setFilter", { filter });
      } catch (error) {
        context.commit("clearListTags");
        throw error;
      }
    },

    async pushTagToDevice(context, { tenant, uid, name }) {
      try {
        await apiTags.pushTagToDevice(tenant, uid, name);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async removeTagFromDevice(context, { tenant, uid, name }) {
      try {
        await apiTags.removeTagFromDevice(tenant, uid, name);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async pushTagToFirewallRule(context, { tenant, name }) {
      try {
        await apiTags.pushTagToFirewallRule(tenant, name);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async removeTagFromFirewallRule(context, { tenant, name }) {
      try {
        await apiTags.removeTagFromFirewallRule(tenant, name);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async pushTagToPublicKey(context, { tenant, fingerprint, name }) {
      try {
        await apiTags.pushTagToPublicKey(tenant, fingerprint, name);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async removeTagFromPublicKey(context, { tenant, fingerprint, name }) {
      try {
        await apiTags.removeTagFromPublicKey(tenant, fingerprint, name);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    setSelected: async (context, data) => {
      context.commit("setSelected", data);
    },

    clearSelectedTags(context) {
      context.commit("clearSelected");
    },
  },
};
