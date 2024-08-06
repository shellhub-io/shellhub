import { Module } from "vuex";

import * as apiContainer from "../api/container";
import { State } from "..";
import { IContainer } from "@/interfaces/IContainer";

export interface ContainerState {
  containers: Array<IContainer>;
  container: IContainer;
  numberContainers: number;
  page: number;
  perPage: number;
  filter: undefined | string;
  status: "accepted" | "rejected" | "pending" | "unused";
  sortStatusField: undefined | string;
  sortStatusString: "asc" | "desc" | "";
  showContainers: boolean;
  }

export const container: Module<ContainerState, State> = {
  namespaced: true,
  state: {
    containers: [],
    container: {} as IContainer,
    numberContainers: 0,
    page: 1,
    perPage: 10,
    filter: "",
    status: "accepted",
    sortStatusField: undefined,
    sortStatusString: "asc",
    showContainers: false,
  },

  getters: {
    list: (state) => state.containers,
    get: (state) => state.container,
    getName: (state) => state.container.name,
    getNumberContainers: (state) => state.numberContainers,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
    getFilter: (state) => state.filter,
    getStatus: (state) => state.status,
    getSortStatusField: (state) => state.sortStatusField,
    getSortStatusString: (state) => state.sortStatusString,
    getShowContainers: (state) => state.showContainers,
  },

  mutations: {
    setContainers: (state, res) => {
      state.containers = res.data;
      state.numberContainers = parseInt(res.headers["x-total-count"], 10);
    },

    setShowContainers: (state) => {
      state.showContainers = true;
    },

    setContainer: (state, data) => {
      state.container = data;
    },

    renameContainer: (state, data) => {
      const { container } = state;
      container.name = data.name;
      state.container = container;
    },

    setPagePerpageFilter: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
      state.filter = data.filter;
      state.status = data.status;
      state.sortStatusField = data.sortStatusField;
      state.sortStatusString = data.sortStatusString;
    },

    setFilter: (state, filter) => {
      state.filter = filter;
    },

    clearListContainers: (state) => {
      state.containers = [];
      state.numberContainers = 0;
    },

    clearObjectContainer: (state) => {
      state.container = {} as IContainer;
    },
  },

  actions: {
    fetch: async ({ commit }, data) => {
      try {
        const res = await apiContainer.fetchContainers(
          data.page,
          data.perPage,
          data.filter,
          data.status,
          data.sortStatusField,
          data.sortStatusString,
        );
        if (res.data.length && data.committable === false) {
          commit("setShowContainers");
          return;
        }
        commit("setContainers", res);
        commit("setPagePerpageFilter", data);
      } catch (error) {
        commit("clearListContainers");
        throw error;
      }
    },

    remove: async (context, uid) => {
      await apiContainer.removeContainer(uid);
    },

    rename: async (context, data) => {
      await apiContainer.renameContainer(data);
      context.commit("renameContainer", data);
    },

    get: async (context, uid) => {
      try {
        const res = await apiContainer.getContainer(uid);
        context.commit("setContainer", res.data);
      } catch (error) {
        context.commit("clearObjectContainer");
        throw error;
      }
    },

    accept: async (context, uid) => {
      try {
        await apiContainer.acceptContainer(uid);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    reject: async (context, uid) => {
      try {
        await apiContainer.rejectContainer(uid);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    setFilter: async (context, filter) => {
      context.commit("setFilter", filter);
    },

    refresh: async ({ commit, state }) => {
      try {
        const res = await apiContainer.fetchContainers(
          state.page,
          state.perPage,
          state.filter,
          state.status,
          state.sortStatusField,
          state.sortStatusString,
        );
        commit("setContainers", res);
      } catch (error) {
        commit("clearListContainers");
        throw error;
      }
    },

    async search({ commit, state }, data) {
      try {
        const res = await apiContainer.fetchContainers(
          data.page,
          data.perPage,
          data.filter,
          state.status,
          state.sortStatusField,
          state.sortStatusString,
        );
        commit("setContainers", res);
        commit("setFilter", data.filter);
      } catch (error) {
        commit("clearListContainers");
        throw error;
      }
    },

    updateDeviceTag: async (context, data) => {
      try {
        await apiContainer.updateContainerTag(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },
  },
};
