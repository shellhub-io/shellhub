import { AxiosError, AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from "axios";
import { router } from "../router";
import useAuthStore from "@/store/modules/auth";
import useSpinnerStore from "@/store/modules/spinner";

const onRequest = (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
  useSpinnerStore().status = true;
  return config;
};

const onRequestError = (error: AxiosError) => {
  useSpinnerStore().status = false;
  return Promise.reject(error);
};

const onResponse = (response: AxiosResponse): AxiosResponse => {
  useSpinnerStore().status = false;
  return response;
};

const onResponseError = async (error: AxiosError): Promise<AxiosError> => {
  const { logout } = useAuthStore();
  useSpinnerStore().status = false;
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
