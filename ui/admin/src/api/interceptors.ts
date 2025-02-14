/* eslint-disable */
// @ts-nocheck
import { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from "axios";
import { store } from "../store";
// import router from "../router/index";
import router from "../router";

const onRequest = (config: AxiosRequestConfig): AxiosRequestConfig => {
  store.dispatch("spinner/setStatus", true);
  return config;
};

const onRequestError = (error: AxiosError): Promise<AxiosError> => {
  store.dispatch("spinner/setStatus", false);
  Promise.reject(error)
};

const onResponse = (response: AxiosResponse): AxiosResponse => {
  store.dispatch("spinner/setStatus", false);
  return response;
};

const onResponseError = async (error: AxiosError): Promise<AxiosError> => {
  store.dispatch("spinner/setStatus", false);
  if (error.response.status === 401) {
    await store.dispatch("auth/logout");
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
