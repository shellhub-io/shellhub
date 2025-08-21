import Axios from "axios";
import { router } from "@/router";
import useAuthStore from "../modules/auth";
import useSpinnerStore from "@/store/modules/spinner";

export default () => {
  const axios = Axios.create({
    baseURL: `${window.location.protocol}//${window.location.host}/api`,
    headers: {
      Authorization: `Bearer ${localStorage.getItem("token")}`,
    },
  });

  const { logout } = useAuthStore();
  const spinnerStore = useSpinnerStore();

  axios.interceptors.request.use(
    (config) => {
      spinnerStore.status = true;
      return config;
    },
    async (error) => {
      throw error;
    },
  );

  axios.interceptors.response.use(
    (response) => {
      spinnerStore.status = false;
      return response;
    },
    async (error) => {
      spinnerStore.status = false;
      if (error.response.status === 401) {
        logout();
        await router.push({ name: "Login" });
      }
      throw error;
    },
  );

  return axios;
};
