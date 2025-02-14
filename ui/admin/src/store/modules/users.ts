import { Module } from "vuex";
import { State } from "../index";
import * as apiUser from "../api/users";
import { IUser } from "../../interfaces/IUser";

export interface UsersState {
  users: IUser[];
  user: IUser;
  filter: string;
  numberUsers: number;
  page: number;
  perPage: number;
  ownedNamespaces: number;
  generatedPassword: string;
}

export const users: Module<UsersState, State> = {
  namespaced: true,
  state: {
    users: [],
    user: {} as IUser,
    filter: "",
    numberUsers: 0,
    page: 0,
    perPage: 0,
    ownedNamespaces: 0,
    generatedPassword: "",
  },
  getters: {
    perPage: (state) => state.perPage,
    users: (state) => state.users,
    user: (state) => state.user,
    page: (state) => state.page,
    filter: (state) => state.filter,
    numberUsers: (state) => state.numberUsers,
    ownedNamespaces: (state) => state.ownedNamespaces,
    getGeneratedPassword: (state) => state.generatedPassword,
  },
  mutations: {
    setUsers: (state, res) => {
      state.users = res.data;
      state.numberUsers = parseInt(res.headers["x-total-count"], 10);
    },

    setUser: (state: UsersState, res) => {
      const { user, namespacesOwned } = res.data;
      state.user = user;
      state.ownedNamespaces = namespacesOwned;
    },

    setPageAndPerPage: (state: UsersState, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
    },

    setUserFilter: (state: UsersState, filter: string) => {
      state.filter = filter;
    },

    setPassword: (state, password) => {
      state.generatedPassword = password;
    },

    clearListUsers: (state: UsersState) => {
      state.users = [];
      state.numberUsers = 0;
    },
  },

  actions: {
    async fetch({ commit }, data) {
      const { page, perPage, filter } = data;
      const res = await apiUser.fetchUsers(perPage, page, filter);
      if (res.data.length) {
        commit("setPageAndPerPage", data);
        commit("setUsers", res);
        commit("setUserFilter", filter);
        return true;
      }

      return false;
    },

    async exportUsersToCsv({ state }) {
      const { data } = await apiUser.exportUsers(state.filter);
      return data;
    },

    async setFilterUsers({ commit }, filter) {
      commit("setUserFilter", filter);
    },

    async addUser(context, data) {
      await apiUser.addUser(data);
    },

    async search({ commit }, data) {
      try {
        const res = await apiUser.fetchUsers(data.perPage, data.page, data.filter);
        commit("setUsers", res);
        commit("setUserFilter", data.filter);
      } catch (error) {
        commit("clearListUsers");
        throw error;
      }
    },

    async get(context, id) {
      const res = await apiUser.getUser(id);
      context.commit("setUser", res);
    },

    async put(context, data) {
      const { id } = data;
      await apiUser.putUser(id, data);
    },

    async remove(context, id) {
      await apiUser.removeUser(id);
    },

    async resetUserPassword({ commit }, id) {
      const res = await apiUser.resetUserPassword(id);
      commit("setPassword", res.data.password);
    },

    async refresh({ commit, state }) {
      try {
        const res = await apiUser.fetchUsers(state.perPage, state.page, state.filter);
        commit("setUsers", res);
        commit("setUserFilter", state.filter);
      } catch (error) {
        commit("clearListUsers");
        throw error;
      }
    },
  },
};
