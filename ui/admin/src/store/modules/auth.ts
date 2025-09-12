import { defineStore } from "pinia";
import { ref, computed } from "vue";
import * as authApi from "../api/auth";

const useAuthStore = defineStore("auth", () => {
  const status = ref<string>("");
  const token = ref<string>(localStorage.getItem("cloud_token") || "");
  const currentUser = ref<string>(localStorage.getItem("cloud_user") || "");
  const isLoggedIn = computed(() => !!token.value);

  const login = async (userData: { username: string; password: string }) => {
    status.value = "loading";

    try {
      const { data } = await authApi.login(userData);
      status.value = "success";
      token.value = data.token ?? "";
      currentUser.value = data.user ?? "";

      localStorage.setItem("cloud_token", data.token || "");
      localStorage.setItem("cloud_user", data.user || "");
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

    localStorage.removeItem("cloud_token");
    localStorage.removeItem("cloud_user");
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
