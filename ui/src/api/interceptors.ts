import { AxiosError, AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from "axios";
import useAdminAuthStore from "@admin/store/modules/auth";
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

const onResponseError = async (error: AxiosError, isAdmin: boolean): Promise<AxiosError> => {
  useSpinnerStore().status = false;
  if (isAdmin) {
    const router = (await import("@admin/router")).default;
    if (error.response?.status === 401) {
      useAdminAuthStore().logout();
      await router.push({ name: "login" });
    } else if (error.response?.status === 402) {
      await router.push({ name: "license" });
    }
  } else if (error.response?.status === 401) {
    const { router } = await import("@/router");
    useAuthStore().logout();
    await router.push({ name: "Login", query: router.currentRoute.value.query });
  }
  return Promise.reject(error);
};

// eslint-disable-next-line import/prefer-default-export
export function setupInterceptorsTo(axiosInstance: AxiosInstance, isAdmin = false): AxiosInstance {
  axiosInstance.interceptors.request.use(onRequest, onRequestError);
  axiosInstance.interceptors.response.use(onResponse, (error) => onResponseError(error, isAdmin));
  return axiosInstance;
}
