import { Module } from "vuex";
import { State } from "./../index";
import { login, getToken } from "../api/auth";

export interface AuthState {
  status: string;
  token: string;
  user: string;
  name: string;
  tenant: string;
}

export const auth: Module<AuthState, State> = {
  namespaced: true,
  state: {
    status: "",
    token: localStorage.getItem("cloud_token") || "",
    user: localStorage.getItem("cloud_user") || "",
    name: localStorage.getItem("name") || "",
    tenant: localStorage.getItem("tenant") || "",
  },
  getters: {
    isLoggedIn: (state) => !!state.token,
    currentUser: (state) => state.user,
    authStatus: (state) => state.status,
    tenant: (state) => state.tenant,
  },

  mutations: {
    authRequest(state) {
      state.status = "loading";
    },

    authSuccess(state, data: AuthState) {
      state.status = "success";
      state.token = data.token;
      state.user = data.user;
      state.name = data.name;
      state.tenant = data.tenant;
    },

    authError(state) {
      state.status = "error";
    },

    logout(state) {
      state.status = "";
      state.token = "";
      state.name = "";
      state.user = "";
      state.tenant = "";
    },
  },

  actions: {
    async login({ commit }, user) {
      commit("authRequest");

      try {
        const { data } = await login(user);
        localStorage.setItem("cloud_token", data.token || "");
        localStorage.setItem("cloud_user", data.user || "");
        commit("authSuccess", data);
      } catch (error) {
        commit("authError");
        throw error;
      }
    },

    async loginToken({ commit }, user) {
      try {
        const resp = await getToken(user.id);
        return resp.data.token;
      } catch (error) {
        commit("authError");
        throw error;
      }
    },

    logout({ commit }) {
      commit("logout");
      localStorage.removeItem("cloud_token");
      localStorage.removeItem("cloud_user");
      localStorage.removeItem("cloud_tenant");
    },
  },
};
