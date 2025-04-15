import { defineStore } from "pinia";
import { UserAdminResponse } from "@admin/api/client";
import * as apiUser from "../api/users";
import { IUser } from "../../interfaces/IUser";

export const useUsersStore = defineStore("users", {
  state: () => ({
    users: [] as UserAdminResponse[],
    user: {} as UserAdminResponse,
    filter: "",
    numberUsers: 0,
    page: 0,
    perPage: 0,
    ownedNamespaces: 0,
    generatedPassword: "",
  }),

  getters: {
    getPerPage: (state) => state.perPage,
    getUsers: (state) => state.users,
    getUser: (state) => state.user,
    getPage: (state) => state.page,
    getFilter: (state) => state.filter,
    getNumberUsers: (state) => state.numberUsers,
    getOwnedNamespaces: (state) => state.ownedNamespaces,
    getGeneratedPassword: (state) => state.generatedPassword,
  },

  actions: {
    async fetch(data: { page: number; perPage: number; filter: string }) {
      const { page, perPage, filter } = data;
      const res = await apiUser.fetchUsers(perPage, page, filter);

      if (res.data.length) {
        this.page = data.page;
        this.perPage = data.perPage;
        this.users = res.data;
        this.numberUsers = parseInt(res.headers["x-total-count"], 10);
        this.filter = filter;
        return true;
      }
      return false;
    },

    async exportUsersToCsv() {
      const { data } = await apiUser.exportUsers(this.filter);
      return data;
    },

    async setFilterUsers(filter: string) {
      this.filter = filter;
    },

    async addUser(data) {
      await apiUser.addUser(data);
    },

    async search(data: { perPage: number; page: number; filter: string }) {
      try {
        const res = await apiUser.fetchUsers(data.perPage, data.page, data.filter);
        this.users = res.data;
        this.numberUsers = parseInt(res.headers["x-total-count"], 10);
        this.filter = data.filter;
      } catch (error) {
        this.clearListUsers();
        throw error;
      }
    },

    async get(id: string) {
      const res = await apiUser.getUser(id);
      const { user, namespacesOwned } = res.data;
      this.user = user as IUser;
      this.ownedNamespaces = namespacesOwned as number;
    },

    async put(data) {
      const { id } = data;
      await apiUser.putUser(id, data);
    },

    async remove(id: string) {
      await apiUser.removeUser(id);
    },

    async resetUserPassword(id: string) {
      const res = await apiUser.resetUserPassword(id);
      this.generatedPassword = res.data.password as string;
    },

    async refresh() {
      try {
        const res = await apiUser.fetchUsers(this.perPage, this.page, this.filter);
        this.users = res.data;
        this.numberUsers = parseInt(res.headers["x-total-count"], 10);
      } catch (error) {
        this.clearListUsers();
        throw error;
      }
    },

    clearListUsers() {
      this.users = [];
      this.numberUsers = 0;
    },
  },
});

export default useUsersStore;
