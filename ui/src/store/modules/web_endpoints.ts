import { Module } from "vuex";
import { State } from "../index";
import * as apiTunnel from "../api/web_endpoints";
import { IWebEndpoints, IWebEndpointsCreate, IWebEndpointsDelete } from "@/interfaces/IWebEndpoints";

export interface WebEndpointsState {
  web_endpoints: Array<IWebEndpoints>;
  page: number;
  perPage: number;
  filter: string;
  sortBy: "created_at" | "updated_at" | "address" | "uid";
  orderBy: "asc" | "desc";
  numberEndpoints: number;
  showWebEndpoints: boolean;
}

export const webEndpoints: Module<WebEndpointsState, State> = {
  namespaced: true,
  state: {
    web_endpoints: [],
    page: 1,
    perPage: 10,
    filter: "",
    sortBy: "uid",
    orderBy: "asc",
    numberEndpoints: 0,
    showWebEndpoints: false,
  },

  getters: {
    listWebEndpoints: (state) => state.web_endpoints,
    getFilter: (state) => state.filter,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
    getTotalCount: (state) => state.numberEndpoints,
    getSortBy: (state) => state.sortBy,
    getOrderBy: (state) => state.orderBy,
    getShowWebEndpoints: (state) => state.showWebEndpoints,
  },

  mutations: {
    setWebEndpoints: (state, res) => {
      state.web_endpoints = res.data;
      state.numberEndpoints = parseInt(res.headers["x-total-count"], 10);
    },

    setShowWebEndpoints: (state) => {
      state.showWebEndpoints = true;
    },

    setPagePerPage: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
      state.filter = data.filter;
      state.sortBy = data.sortBy;
      state.orderBy = data.orderBy;
    },

    setFilter: (state, filter) => {
      state.filter = filter;
    },

    setSortStatus: (state, data) => {
      state.sortBy = data.sortBy;
      state.orderBy = data.orderBy;
    },

    clearListEndpoints: (state) => {
      state.web_endpoints = [];
      state.numberEndpoints = 0;
      state.showWebEndpoints = false;
    },
  },

  actions: {
    async get({ commit }, data) {
      try {
        const res = await apiTunnel.getWebEndpoints(
          data.filter,
          data.page,
          data.perPage,
          data.sortBy,
          data.orderBy,
        );
        if (res.data.length) {
          commit("setShowWebEndpoints");
        }
        commit("setWebEndpoints", res);
        commit("setPagePerPage", data);
      } catch (error) {
        commit("clearListEndpoints");
        commit("setShowWebEndpoints");
        throw error;
      }
    },

    async search({ commit, state }, data) {
      try {
        const res = await apiTunnel.getWebEndpoints(
          data.filter,
          data.page,
          data.perPage,
          state.sortBy,
          state.orderBy,
        );
        commit("setWebEndpoints", res);
        commit("setFilter", data.filter);
      } catch (error) {
        commit("clearListEndpoints");
        throw error;
      }
    },

    async delete(_, data: IWebEndpointsDelete) {
      const { address } = data;
      const res = await apiTunnel.deleteWebEndpoints(address);
      return res;
    },

    async create(_, data: IWebEndpointsCreate) {
      const { uid, host, port, ttl } = data;
      const res = await apiTunnel.createWebEndpoints(uid, host, port, ttl);
      return res;
    },
  },
};
