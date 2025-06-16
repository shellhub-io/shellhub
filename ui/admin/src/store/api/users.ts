import { IUser } from "@admin/interfaces/IUser";
import { UserAdminRequest } from "@admin/api/client";
import { adminApi } from "@admin/api/http";

const fetchUsers = async (
  perPage: number,
  page: number,
  search?: string,
) => adminApi.getUsers(search, page, perPage);

const getUser = (id: string) => adminApi.getUser(id);

const exportUsers = async (filter: string) => adminApi.exportUsers(filter);

const addUser = (userData: IUser) => adminApi.createUserAdmin({
  name: userData.name,
  email: userData.email,
  username: userData.username,
  password: userData.password,
  max_namespaces: userData.max_namespaces,
});

const putUser = async (id: string, userData: IUser) => adminApi.adminUpdateUser(id, {
  name: userData.name,
  email: userData.email,
  username: userData.username,
  password: userData.password,
  status: userData.status,
  max_namespaces: userData.max_namespaces,
} as UserAdminRequest);

const resetUserPassword = async (id: string) => adminApi.adminResetUserPassword(id);

const removeUser = (id: string) => adminApi.deleteUser(id);

export { fetchUsers, getUser, exportUsers, addUser, putUser, resetUserPassword, removeUser };
