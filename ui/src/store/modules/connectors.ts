import { Module } from "vuex";
import { State } from "..";
import { IConnector } from "@/interfaces/IConnector";
import * as apiConnector from "../api/connectors";

export interface ConnectorState {
  connectors: Array<IConnector>;
  connector: IConnector;
  info: object;
  page: number;
  perPage: number;
  numberConnectors: number;
}

export const connectors: Module<ConnectorState, State> = {
  namespaced: true,
  state: {
    connectors: [],
    connector: {} as IConnector,
    info: {},
    page: 1,
    perPage: 10,
    numberConnectors: 0,
  },

  getters: {
    list: (state) => state.connectors,
    get: (state) => state.connector,
    getInfo: (state) => state.info,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
    getNumberConnectors: (state) => state.numberConnectors,
  },

  mutations: {
    setConnectors: (state, res) => {
      state.connectors = res.data;
      state.numberConnectors = parseInt(res.headers["x-total-count"], 10);
    },
    setPagePerpage: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
    },
    setConnector: (state, data) => {
      state.connector = data;
    },
    setInfoConnector: (state, data) => {
      state.info = data;
    },
    clearListConnector: (state) => {
      state.connectors = [];
      state.numberConnectors = 0;
    },
    clearConnector: (state) => {
      state.connector = {} as IConnector;
      state.info = {};
    },
  },

  actions: {
    fetch: async ({ commit }, data) => {
      try {
        const res = await apiConnector.listConnector(
          data.enable,
          data.page,
          data.perPage,
        );
        if (res.data.length) {
          commit("setConnectors", res);
          commit("setPagePerpage", data);
          return res;
        }

        commit("clearListConnector");
        return false;
      } catch (error) {
        commit("clearListConnector");
        throw error;
      }
    },
    get: async (context, uid) => {
      const res = await apiConnector.getConnector(uid);
      context.commit("setConnector", res.data);
    },
    getConnectorInfo: async (context, uid) => {
      const res = await apiConnector.getConnectorInfo(uid);
      context.commit("setInfoConnector", res.data);
    },
    post: async (context, data) => {
      await apiConnector.createConnector(data);
    },
    edit: async (context, data) => {
      await apiConnector.updateConnector(data);
    },
    remove: async (context, data) => {
      await apiConnector.deleteConnector(data);
    },
    setStatus: async (context, status) => {
      context.commit("setStatus", status);
    },
  },
};
