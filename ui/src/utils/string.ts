import { HostnameFilter } from "../interfaces/IFilter";

export const displayOnlyTenCharacters = (str: string) => str.length > 10 ? `${str.slice(0, 10)}...` : str;

export const capitalizeText = (str: string) => str.charAt(0).toUpperCase() + str.slice(1);

export const formatSourceIP = (ip: string) => (ip === ".*" ? "Any IP" : ip);

export const formatUsername = (username: string) => username === ".*" ? "All users" : username;

export const formatHostnameFilter = (filter: HostnameFilter) => filter.hostname === ".*" ? "All devices" : filter.hostname;
