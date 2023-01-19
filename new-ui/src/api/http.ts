/* eslint-disable */
// @ts-nocheck
import * as axiosTs from "./client";
import { Configuration } from "./client";
import { setupInterceptorsTo } from "./interceptors";

const configuration = new Configuration();
configuration.basePath = `${window.location.protocol}//${window.location.host}`;
// configuration.basePath = `${window.location.protocol}//localhost:4010`;
configuration.accessToken = localStorage.getItem("token") || "";

let sessionsApi = new axiosTs.SessionsApi(configuration);
let devicesApi = new axiosTs.DevicesApi(configuration);
let defaultApi = new axiosTs.DefaultApi(configuration);
let namespacesApi = new axiosTs.NamespacesApi(configuration);
let sshApi = new axiosTs.SshApi(configuration);
let tagsApi = new axiosTs.TagsApi(configuration);
let usersApi = new axiosTs.UsersApi(configuration);
let billingApi = new axiosTs.BillingApi(configuration);
let rulesApi = new axiosTs.RulesApi(configuration);
let announcementApi = new axiosTs.AdminApi(configuration);

export const createNewClient = () => {
  const newConfiguration = new Configuration();
  newConfiguration.basePath = `${window.location.protocol}//${window.location.host}`;
  // newConfiguration.basePath = `${window.location.protocol}//localhost:4010`;
  newConfiguration.accessToken = localStorage.getItem("token") || "";

  sessionsApi = new axiosTs.SessionsApi(newConfiguration);
  devicesApi = new axiosTs.DevicesApi(newConfiguration);
  defaultApi = new axiosTs.DefaultApi(newConfiguration);
  namespacesApi = new axiosTs.NamespacesApi(newConfiguration);
  sshApi = new axiosTs.SshApi(newConfiguration);
  tagsApi = new axiosTs.TagsApi(newConfiguration);
  usersApi = new axiosTs.UsersApi(newConfiguration);
  billingApi = new axiosTs.BillingApi(newConfiguration);
  rulesApi = new axiosTs.RulesApi(newConfiguration);
  announcementApi = new axiosTs.AdminApi(newConfiguration);
  return newConfiguration;
};

setupInterceptorsTo(sessionsApi.axios);
setupInterceptorsTo(devicesApi.axios);
setupInterceptorsTo(defaultApi.axios);
setupInterceptorsTo(namespacesApi.axios);
setupInterceptorsTo(sshApi.axios);
setupInterceptorsTo(tagsApi.axios);
setupInterceptorsTo(usersApi.axios);
setupInterceptorsTo(billingApi.axios);
setupInterceptorsTo(rulesApi.axios);
setupInterceptorsTo(announcementApi.axios);

export {
  sessionsApi,
  devicesApi,
  defaultApi,
  namespacesApi,
  sshApi,
  tagsApi,
  usersApi,
  billingApi,
  rulesApi,
  announcementApi,
};
