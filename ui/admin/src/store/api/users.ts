import { adminApi } from "./../../api/http";

type userDataType = {
  name: string;
  email: string;
  username: string;
  password: string;
  confirmed?: boolean;
  max_namespaces?: number;
};

const fetchUsers = async (
  perPage: number,
  page: number,
  search?: string,
) => adminApi.getUsers(search, page, perPage);

const getUser = (id: string) => adminApi.getUser(id);

const exportUsers = async (filter: string) => adminApi.exportUsers(filter);

const addUser = (userData: userDataType) => adminApi.createUserAdmin({
  name: userData.name,
  email: userData.email,
  username: userData.username,
  password: userData.password,
  max_namespaces: userData.max_namespaces,
});

const putUser = async (id: string, userData: userDataType) => adminApi.adminUpdateUser(id, {
  name: userData.name,
  email: userData.email,
  username: userData.username,
  password: userData.password,
  confirmed: userData.confirmed,
  max_namespaces: userData.max_namespaces,
});

const resetUserPassword = async (id: string) => adminApi.adminResetUserPassword(id);

const removeUser = (id: string) => adminApi.deleteUser(id);

export { fetchUsers, getUser, exportUsers, addUser, putUser, resetUserPassword, removeUser };
