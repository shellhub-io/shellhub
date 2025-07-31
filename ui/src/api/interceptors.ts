import { AxiosError, AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from "axios";
import { store } from "../store";
import { router } from "../router";
import useAuthStore from "@/store/modules/auth";

const onRequest = (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
  store.dispatch("spinner/setStatus", true);
  return config;
};

const onRequestError = (error: AxiosError) => {
  store.dispatch("spinner/setStatus", false);
  return Promise.reject(error);
};

const onResponse = (response: AxiosResponse): AxiosResponse => {
  store.dispatch("spinner/setStatus", false);
  return response;
};

const onResponseError = async (error: AxiosError): Promise<AxiosError> => {
  const { logout } = useAuthStore();
  store.dispatch("spinner/setStatus", false);
  if (error.response?.status === 401) {
    logout();
    await router.push({ name: "Login", query: router.currentRoute.value.query });
  }
  return Promise.reject(error);
};

// eslint-disable-next-line import/prefer-default-export
export function setupInterceptorsTo(axiosInstance: AxiosInstance): AxiosInstance {
  axiosInstance.interceptors.request.use(onRequest, onRequestError);
  axiosInstance.interceptors.response.use(onResponse, onResponseError);
  return axiosInstance;
}
