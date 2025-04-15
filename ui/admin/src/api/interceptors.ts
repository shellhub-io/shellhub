/* eslint-disable */
// @ts-nocheck
import { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from "axios";
import { setActivePinia, createPinia } from "pinia";
import useSpinnerStore from "@admin/store/modules/spinner";
import useAuthStore from "@admin/store/modules/auth";
import { store } from "../store";
import router from "../router/index";

setActivePinia(createPinia());
const spinnerStore = useSpinnerStore();
const authStore = useAuthStore();


const onRequest = (config: AxiosRequestConfig): AxiosRequestConfig => {
  spinnerStore.setStatus(true);
  return config;
};

const onRequestError = (error: AxiosError): Promise<AxiosError> => {
  spinnerStore.setStatus(false);
  Promise.reject(error)
};

const onResponse = (response: AxiosResponse): AxiosResponse => {
  spinnerStore.setStatus(false);
  return response;
};

const onResponseError = async (error: AxiosError): Promise<AxiosError> => {
  spinnerStore.setStatus(false);
  if (error.response.status === 401) {
    await authStore.logout();
    await router.push({ name: "login" });
  } else if (error.response.status === 402) {
    await router.push({ name: "license" });
  }
  return Promise.reject(error);
};

export function setupInterceptorsTo(axiosInstance: AxiosInstance): AxiosInstance {
  axiosInstance.interceptors.request.use(onRequest, onRequestError);
  axiosInstance.interceptors.response.use(onResponse, onResponseError);
  return axiosInstance;
}
