import { AxiosError, AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import useAuthStore from "@admin/store/modules/auth";
import useSpinnerStore from "@/store/modules/spinner";
import router from "../router/index";

const onRequest = (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
  useSpinnerStore().status = true;
  return config;
};

const onRequestError = (error: AxiosError): Promise<AxiosError> => {
  useSpinnerStore().status = false;
  return Promise.reject(error);
};

const onResponse = (response: AxiosResponse): AxiosResponse => {
  useSpinnerStore().status = false;
  return response;
};

const onResponseError = async (error: AxiosError): Promise<AxiosError> => {
  useSpinnerStore().status = false;
  if (error.response?.status === 401) {
    useAuthStore().logout();
    await router.push({ name: "login" });
  } else if (error.response?.status === 402) {
    await router.push({ name: "license" });
  }
  return Promise.reject(error);
};

// eslint-disable-next-line import/prefer-default-export
export function setupInterceptorsTo(axiosInstance: AxiosInstance): AxiosInstance {
  axiosInstance.interceptors.request.use(onRequest, onRequestError);
  axiosInstance.interceptors.response.use(onResponse, onResponseError);
  return axiosInstance;
}
