import { Module } from "vuex";
import { State } from "./../index";
import * as apiAuth from "../api/auth";

export interface AuthState {
  status: string;
  token: string;
  user: string;
  name: string;
  tenant: string;
  email: string;
  id: string;
  role: string;
}

export const auth: Module<AuthState, State> = {
  namespaced: true,
  state: {
    status: "",
    token: localStorage.getItem("token") || "",
    user: localStorage.getItem("user") || "",
    name: localStorage.getItem("name") || "",
    tenant: localStorage.getItem("tenant") || "",
    email: localStorage.getItem("email") || "",
    id: localStorage.getItem("id") || "",
    role: localStorage.getItem("role") || "",
  },

  getters: {
    isLoggedIn: (state) => !!state.token,
    authStatus: (state) => state.status,
    currentUser: (state) => state.user,
    currentName: (state) => state.name,
    tenant: (state) => state.tenant,
    email: (state) => state.email,
    id: (state) => state.id,
    role: (state) => state.role,
  },

  mutations: {
    authRequest(state) {
      state.status = "loading";
    },

    authSuccess(state, data) {
      state.status = "success";
      state.token = data.token;
      state.user = data.user;
      state.name = data.name;
      state.tenant = data.tenant;
      state.email = data.email;
      state.id = data.id;
      state.role = data.role;
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
      state.email = "";
      state.role = "";
    },

    changeData(state, data) {
      state.name = data.name;
      state.user = data.username;
      state.email = data.email;
    },
  },

  actions: {
    async login(context, user) {
      context.commit("authRequest");

      try {
        const resp = await apiAuth.login(user);

        localStorage.setItem("token", resp.data.token || "");
        localStorage.setItem("user", resp.data.user || "");
        localStorage.setItem("name", resp.data.name || "");
        localStorage.setItem("tenant", resp.data.tenant || "");
        localStorage.setItem("email", resp.data.email || "");
        localStorage.setItem("id", resp.data.id || "");
        localStorage.setItem("namespacesWelcome", JSON.stringify({}));
        localStorage.setItem("role", resp.data.role || "");

        context.commit("authSuccess", resp.data);
      } catch (error) {
        context.commit("authError");
        throw error;
      }
    },

    async loginToken(context, token) {
      context.commit("authRequest");

      localStorage.setItem("token", token);

      try {
        const resp = await apiAuth.info();

        localStorage.setItem("user", resp.data.user);
        localStorage.setItem("name", resp.data.name);
        localStorage.setItem("tenant", resp.data.tenant);
        localStorage.setItem("id", resp.data.id);
        localStorage.setItem("email", resp.data.email);
        localStorage.setItem("namespacesWelcome", JSON.stringify({}));
        localStorage.setItem("role", resp.data.role);

        context.commit("authSuccess", resp.data);
      } catch (error) {
        context.commit("authError");
        throw error;
      }
    },

    logout(context) {
      context.commit("logout");
      localStorage.removeItem("token");
      localStorage.removeItem("user");
      localStorage.removeItem("tenant");
      localStorage.removeItem("namespacesWelcome");
      localStorage.removeItem("noNamespace");
      localStorage.removeItem("email");
      localStorage.removeItem("id");
      localStorage.removeItem("name");
      localStorage.removeItem("role");
    },

    changeUserData(context, data) {
      localStorage.setItem("name", data.name);
      localStorage.setItem("user", data.username);
      localStorage.setItem("email", data.email);
      context.commit("changeData", data);
    },

    setShowWelcomeScreen(context, tenantID : string) {
      localStorage.setItem("namespacesWelcome", JSON.stringify(
        Object.assign(
          JSON.parse(localStorage.getItem("namespacesWelcome") || "") || {},
          { ...{ [tenantID]: true } },
        ),
      ));
    },
  },
};
