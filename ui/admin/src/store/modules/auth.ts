import { defineStore } from "pinia";
import { ref, computed } from "vue";
import * as authApi from "../api/auth";

const useAuthStore = defineStore("auth", () => {
  const status = ref<string>("");
  const token = ref<string>(localStorage.getItem("token") || "");
  const currentUser = ref<string>(localStorage.getItem("user") || "");
  const isLoggedIn = computed(() => !!token.value);

  const login = async (userData: { username: string; password: string }) => {
    status.value = "loading";

    try {
      const { data } = await authApi.login(userData);
      status.value = "success";
      token.value = data.token ?? "";
      currentUser.value = data.user ?? "";

      localStorage.setItem("token", data.token || "");
      localStorage.setItem("user", data.user || "");
    } catch (error) {
      status.value = "error";
      throw error;
    }
  };

  const getLoginToken = async (userId: string) => {
    try {
      const resp = await authApi.getToken(userId);
      return resp.data.token;
    } catch (error) {
      status.value = "error";
      throw error;
    }
  };

  const logout = () => {
    status.value = "";
    token.value = "";
    currentUser.value = "";

    localStorage.removeItem("token");
    localStorage.removeItem("user");
  };

  return {
    status,
    token,
    currentUser,
    isLoggedIn,
    login,
    getLoginToken,
    logout,
  };
});

export default useAuthStore;
