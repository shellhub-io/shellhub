import * as axiosTs from "./client";
import { Configuration } from "./client";
import { setupInterceptorsTo } from "./interceptors";

// This is the default configuration for local instance endpoints
const configuration = new Configuration();
configuration.basePath = `${window.location.protocol}//${window.location.host}`;
configuration.accessToken = localStorage.getItem("token") || "";

// We need a custom configuration for cloud endpoints
const cloudApiConfiguration = new Configuration();
cloudApiConfiguration.basePath = "https://cloud.shellhub.io";

const sessionsApi = new axiosTs.SessionsApi(configuration);
const devicesApi = new axiosTs.DevicesApi(configuration);
const defaultApi = new axiosTs.DefaultApi(configuration);
const namespacesApi = new axiosTs.NamespacesApi(configuration);
const sshApi = new axiosTs.SshApi(configuration);
const tagsApi = new axiosTs.TagsApi(configuration);
const usersApi = new axiosTs.UsersApi(configuration);
const billingApi = new axiosTs.BillingApi(configuration);
const rulesApi = new axiosTs.RulesApi(configuration);
const announcementApi = new axiosTs.AnnouncementsApi(cloudApiConfiguration);

/**
 * @deprecated This method is deprecated and no longer performs any action,
 * kept for backward compatibility but it will be removed in the future.
**/
export const createNewClient = () => Function;

[
  sessionsApi,
  devicesApi,
  defaultApi,
  namespacesApi,
  sshApi,
  tagsApi,
  usersApi,
  billingApi,
  rulesApi,
].forEach((instance) => {
  /* eslint-disable */
  // @ts-ignore Ignore next line because 'axios' is a protected property
  setupInterceptorsTo(instance.axios);
  /* eslint-enable */
});

export {
  configuration,
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
