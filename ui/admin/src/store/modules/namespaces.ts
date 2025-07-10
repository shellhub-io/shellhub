import { defineStore } from "pinia";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import * as apiNamespace from "../api/namespaces";
import { useUsersStore } from "./users";

export const useNamespacesStore = defineStore("namespace", {
  state: () => ({
    namespaces: [] as IAdminNamespace[],
    namespace: {} as IAdminNamespace,
    perPage: 0,
    page: 0,
    filter: "",
    numberNamespaces: 0,
  }),

  getters: {
    list: (state) => state.namespaces,
    getNamespace: (state) => state.namespace,
    getnumberOfNamespaces: (state) => state.numberNamespaces,
    getFilter: (state) => state.filter,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
  },

  actions: {
    setNamespaces(res) {
      this.namespaces = res.data;
      this.numberNamespaces = parseInt(res.headers["x-total-count"], 10);
    },

    setNamespace(res) {
      this.namespace = res.data;
    },

    setNamespaceFilter(filter: string) {
      this.filter = filter;
    },

    setPageAndPerPage(data: { page: number; perPage: number }) {
      this.page = data.page;
      this.perPage = data.perPage;
    },

    clearListNamespaces() {
      this.namespaces = [];
      const usersStore = useUsersStore();
      usersStore.clearListUsers?.();
    },

    async fetch(data: { page: number; perPage: number; filter: string }) {
      const res = await apiNamespace.fetchNamespaces(data.page, data.perPage, data.filter);
      if (res.data.length) {
        this.setPageAndPerPage({ perPage: data.perPage, page: data.page });
        this.setNamespaces(res);
        this.setNamespaceFilter(data.filter);
        return true;
      }
      return false;
    },

    async get(id: string) {
      const res = await apiNamespace.getNamespace(id);
      this.setNamespace(res);
    },

    async exportNamespacesToCsv() {
      const { data } = await apiNamespace.exportNamespaces(this.filter);
      return data;
    },

    async setFilterNamespaces(filter: string) {
      this.setNamespaceFilter(filter);
    },

    async refresh() {
      try {
        const res = await apiNamespace.fetchNamespaces(this.page, this.perPage, this.filter);
        this.setNamespaces(res);
      } catch (error) {
        this.clearListNamespaces();
        throw error;
      }
    },

    async search(data: { perPage: number; page: number; filter: string }) {
      try {
        const res = await apiNamespace.fetchNamespaces(data.perPage, data.page, data.filter);
        this.setNamespaces(res);
        this.setNamespaceFilter(data.filter);
      } catch (error) {
        this.clearListNamespaces();
        throw error;
      }
    },

    async put(data: IAdminNamespace) {
      await apiNamespace.updateNamespace(data);
    },
  },
});

export default useNamespacesStore;
