// stores/auth.ts
import { defineStore } from "pinia";
import { login, getToken } from "../api/auth";

export interface AuthState {
  status: string;
  token: string;
  user: string;
}

export const useAuthStore = defineStore("auth", {
  state: (): AuthState => ({
    status: "",
    token: localStorage.getItem("cloud_token") || "",
    user: localStorage.getItem("cloud_user") || "",
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    currentUser: (state) => state.user,
    authStatus: (state) => state.status,
  },

  actions: {
    async login(user: { username: string; password: string }) {
      this.status = "loading";

      try {
        const { data } = await login(user);
        this.status = "success";
        this.token = data.token ?? "";
        this.user = data.user ?? "";

        localStorage.setItem("cloud_token", data.token || "");
        localStorage.setItem("cloud_user", data.user || "");
      } catch (error) {
        this.status = "error";
        throw error;
      }
    },

    async loginToken(user: { id: string }) {
      try {
        const resp = await getToken(user.id);
        return resp.data.token;
      } catch (error) {
        this.status = "error";
        throw error;
      }
    },

    logout() {
      this.status = "";
      this.token = "";
      this.user = "";

      localStorage.removeItem("cloud_token");
      localStorage.removeItem("cloud_user");
    },
  },
});

export default useAuthStore;
