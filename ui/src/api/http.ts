import axios, { AxiosInstance } from "axios";
import * as axiosTs from "./client";
import { Configuration } from "./client";
import { BaseAPI } from "./client/base";
import { setupInterceptorsTo } from "./interceptors";

// This is the default configuration for local instance endpoints
const configuration = new Configuration();
configuration.basePath = `${window.location.protocol}//${window.location.host}`;
configuration.accessToken = localStorage.getItem("token") || "";

// We need a custom configuration for cloud endpoints
const cloudApiConfiguration = new Configuration();
cloudApiConfiguration.basePath = "https://cloud.shellhub.io";

// Creates a new axios instance and setup interceptors by default
const newAxiosInstance = (setupInterceptor = true): AxiosInstance => {
  const instance = axios.create();
  if (setupInterceptor) setupInterceptorsTo(instance);
  return instance;
};
const sessionsApi = new axiosTs.SessionsApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const devicesApi = new axiosTs.DevicesApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const defaultApi = new axiosTs.DefaultApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const namespacesApi = new axiosTs.NamespacesApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const apiKeysApi = new axiosTs.ApiKeysApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const sshApi = new axiosTs.SshApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const tagsApi = new axiosTs.TagsApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const usersApi = new axiosTs.UsersApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const mfaApi = new axiosTs.MfaApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const billingApi = new axiosTs.BillingApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const rulesApi = new axiosTs.RulesApi(
  configuration,
  undefined,
  newAxiosInstance(),
);
const announcementApi = new axiosTs.AnnouncementsApi(
  cloudApiConfiguration,
  undefined,
  newAxiosInstance(false),
);

/**
 * @deprecated This method is deprecated and no longer performs any action,
 * kept for backward compatibility but it will be removed in the future.
 **/
export const createNewClient = () => Function;

/**
 * Extends the interface BaseAPI to include a new method `getAxios` allowing
 * access to the protected axios property without the need to access it
 * directly from outside the class and avoiding a linting error caused by
 * accessing a protected property.
 **/
declare module "./client/base" {
  interface BaseAPI {
    getAxios(): AxiosInstance;
  }
}

/** Returns the axios instance */
BaseAPI.prototype.getAxios = function getAxios(this: BaseAPI): AxiosInstance {
  return this.axios;
};

export {
  configuration,
  sessionsApi,
  devicesApi,
  defaultApi,
  namespacesApi,
  apiKeysApi,
  sshApi,
  tagsApi,
  usersApi,
  mfaApi,
  billingApi,
  rulesApi,
  announcementApi,
};
