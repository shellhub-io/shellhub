import { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from "axios";
import { store } from "../store";
import router from "../router";

const onRequest = (config: AxiosRequestConfig): AxiosRequestConfig => {
  store.dispatch("spinner/setStatus", true);
  return config;
};

const onRequestError = (error: AxiosError): any => {
  store.dispatch("spinner/setStatus", false);
  Promise.reject(error);
};

const onResponse = (response: AxiosResponse): AxiosResponse => {
  store.dispatch("spinner/setStatus", false);
  return response;
};

const onResponseError = async (error: AxiosError): Promise<AxiosError> => {
  store.dispatch("spinner/setStatus", false);
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  if (error.response.status === 401) {
    await store.dispatch("auth/logout");
    await router.push({ name: "login" });
    store.dispatch("layout/setLayout", "simpleLayout");
  }
  return Promise.reject(error);
};

// eslint-disable-next-line import/prefer-default-export
export function setupInterceptorsTo(axiosInstance: AxiosInstance): AxiosInstance {
  axiosInstance.interceptors.request.use(onRequest, onRequestError);
  axiosInstance.interceptors.response.use(onResponse, onResponseError);
  return axiosInstance;
}
