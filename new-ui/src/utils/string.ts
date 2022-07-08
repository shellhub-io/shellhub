import { filterType } from "../interfaces/IFirewallRule";

export const displayOnlyTenCharacters = (str: string) => {
  if (str !== undefined) {
    if (str.length > 10) return `${str.slice(0, 10)}...`;
  }
  return str;
};

export const capitalizeText = (str: string) => {
  if (str !== undefined) {
    return str.charAt(0).toUpperCase() + str.slice(1);
  }
  return str;
};

export const formatSourceIP = (ip: string) => (ip === ".*" ? "Any IP" : ip);

export const formatUsername = (username: string) =>
  username === ".*" ? "All users" : username;

export const formatHostnameFilter = (filter: filterType) =>
  filter.hostname === ".*" ? "All devices" : filter.hostname;
