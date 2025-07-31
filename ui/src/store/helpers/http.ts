import Axios from "axios";
import { router } from "@/router";
import { store } from "..";
import useAuthStore from "../modules/auth";

export default () => {
  const axios = Axios.create({
    baseURL: `${window.location.protocol}//${window.location.host}/api`,
    headers: {
      Authorization: `Bearer ${localStorage.getItem("token")}`,
    },
  });

  const { logout } = useAuthStore();

  axios.interceptors.request.use(
    (config) => {
      store.dispatch("spinner/setStatus", true);
      return config;
    },
    async (error) => {
      throw error;
    },
  );

  axios.interceptors.response.use(
    (response) => {
      store.dispatch("spinner/setStatus", false);
      return response;
    },
    async (error) => {
      store.dispatch("spinner/setStatus", false);
      if (error.response.status === 401) {
        logout();
        await router.push({ name: "Login" });
      }
      throw error;
    },
  );

  return axios;
};
