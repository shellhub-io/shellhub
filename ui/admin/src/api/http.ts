import { AxiosInstance } from "axios";
import { setupInterceptorsTo } from "./interceptors";
import * as axiosTs from "./client";
import { Configuration } from "./client";
import { BaseAPI } from "./client/base";

const configuration = new Configuration();
configuration.basePath = `${window.location.protocol}//${window.location.host}`;
configuration.accessToken = localStorage.getItem("token") || "";

// eslint-disable-next-line import/no-mutable-exports
let adminApi = new axiosTs.AdminApi(configuration);
// eslint-disable-next-line import/no-mutable-exports
let cloudApi = new axiosTs.CloudApi(configuration);

export const createNewClient = () => {
  const newConfiguration = new Configuration();
  newConfiguration.basePath = `${window.location.protocol}//${window.location.host}`;
  newConfiguration.accessToken = localStorage.getItem("token") || "";
  adminApi = new axiosTs.AdminApi(newConfiguration);
  cloudApi = new axiosTs.CloudApi(newConfiguration);
  return { adminApi, cloudApi };
};

declare module "./client/base" {
  interface BaseAPI {
    getAxios(): AxiosInstance;
    getConfiguration(): Configuration | undefined;
    setConfiguration(configuration: Configuration): void;
  }
}

/** Returns the axios instance */
BaseAPI.prototype.getAxios = function getAxios(this: BaseAPI): AxiosInstance {
  return this.axios;
};

/** Returns the configuration */
// eslint-disable-next-line vue/max-len
BaseAPI.prototype.getConfiguration = function getConfiguration(this: BaseAPI): Configuration | undefined {
  return this.configuration;
};

/** Sets the configuration */
// eslint-disable-next-line vue/max-len
BaseAPI.prototype.setConfiguration = function setConfiguration(this: BaseAPI, configuration: Configuration): void {
  this.configuration = configuration;
};

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
setupInterceptorsTo(adminApi.axios);

export { adminApi, cloudApi };
