import { Module } from "vuex";
import { State } from "./../index";
import * as apiNamespace from "../api/namespaces";
import { store } from "../index";
import { INamespace } from "./../../interfaces/INamespace";

export interface NamespacesState {
  namespaces: Array<INamespace>;
  namespace: INamespace;
  perPage: number;
  page: number;
  filter: string;
  numberNamespaces: number;
}

export const namespaces: Module<NamespacesState, State> = {
  namespaced: true,

  state: {
    namespaces: [],
    namespace: {} as INamespace,
    perPage: 0,
    page: 0,
    filter: "",
    numberNamespaces: 0,
  },

  getters: {
    list: (state) => state.namespaces,
    get: (state) => state.namespace,
    numberOfNamespaces: (state) => state.numberNamespaces,
    getFilter: (state) => state.filter,
    page: (state) => state.page,
    perPage: (state) => state.perPage,
  },

  mutations: {
    setNamespaces: (state, res) => {
      state.namespaces = res.data;
      state.numberNamespaces = parseInt(res.headers["x-total-count"], 10);
    },

    setNamespace: (state, res) => {
      state.namespace = res.data;
    },

    setNamespaceFilter: (state, filter) => {
      state.filter = filter;
    },

    setPageAndPerPage: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
    },

    clearListNamespaces: (state) => {
      state.namespaces = [];
      store.commit("users/clearListUsers");
    },
  },

  actions: {
    async fetch({ commit, state }, data: NamespacesState) {
      const res = await apiNamespace.fetchNamespaces(data.page, data.perPage, data.filter);
      if (res.data.length) {
        commit("setPageAndPerPage", {
          perPage: state.perPage,
          page: state.page,
        });
        commit("setNamespaces", res);
        commit("setNamespaceFilter", data.filter);
        return true;
      }

      return false;
    },

    async get({ commit }, id) {
      const res = await apiNamespace.getNamespace(id);
      commit("setNamespace", res);
    },

    async exportNamespacesToCsv({ state }) {
      const { data } = await apiNamespace.exportNamespaces(state.filter);
      return data;
    },

    async setFilterNamespaces({ commit }, filter) {
      commit("setNamespaceFilter", filter);
    },

    async refresh({ commit, state }) {
      try {
        const res = await apiNamespace.fetchNamespaces(state.perPage, state.page, state.filter);
        commit("setNamespaces", res);
      } catch (error) {
        commit("clearListNamespaces");
        throw error;
      }
    },

    async search({ commit }, data) {
      try {
        const res = await apiNamespace.fetchNamespaces(data.perPage, data.page, data.filter);
        commit("setNamespaces", res);
        commit("setNamespaceFilter", data.filter);
      } catch (error) {
        commit("clearListNamespaces");
        throw error;
      }
    },

    async put(context, data: INamespace) {
      await apiNamespace.updateNamespace(data);
    },
  },
};
