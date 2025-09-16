import { IAdminUserFormData } from "@admin/interfaces/IUser";
import { UserAdminRequest } from "@admin/api/client";
import { adminApi } from "@admin/api/http";

export const fetchUsers = async (
  page: number,
  perPage: number,
  search?: string,
) => adminApi.getUsers(search, page, perPage);

export const getUser = (id: string) => adminApi.getUser(id);

export const exportUsers = async (filter: string) => adminApi.exportUsers(filter);

export const addUser = (userData: IAdminUserFormData) => adminApi.createUserAdmin({
  name: userData.name,
  email: userData.email,
  username: userData.username,
  password: userData.password,
  max_namespaces: userData.max_namespaces,
});

export const updateUser = async (id: string, userData: IAdminUserFormData) => adminApi.adminUpdateUser(id, {
  name: userData.name,
  email: userData.email,
  username: userData.username,
  password: userData.password,
  status: userData.status,
  max_namespaces: userData.max_namespaces,
} as UserAdminRequest);

export const resetUserPassword = async (id: string) => adminApi.adminResetUserPassword(id);

export const deleteUser = (id: string) => adminApi.deleteUser(id);
