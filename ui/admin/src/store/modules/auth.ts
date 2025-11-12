import { defineStore } from "pinia";
import { ref, computed } from "vue";
import * as authApi from "../api/auth";

const useAuthStore = defineStore("auth", () => {
  const status = ref<string>("");
  const token = ref<string>(localStorage.getItem("token") || "");
  const currentUser = ref<string>(localStorage.getItem("user") || "");
  const isAdmin = ref<boolean>(localStorage.getItem("admin") === "true");
  const isLoggedIn = computed(() => !!token.value);

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
    isAdmin,
    isLoggedIn,
    getLoginToken,
    logout,
  };
});

export default useAuthStore;
