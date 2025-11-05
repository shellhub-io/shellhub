import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminUser, IAdminUserFormData } from "@admin/interfaces/IUser";
import * as usersApi from "../api/users";

const useUsersStore = defineStore("users", () => {
  const users = ref<IAdminUser[]>([]);
  const usersCount = ref<number>(0);

  const currentFilter = ref<string>("");

  const setFilter = (filter: string) => {
    currentFilter.value = filter || "";
  };

  const fetchUsersList = async (data?: { page?: number; perPage?: number; filter?: string }) => {
    const filter = data?.filter ?? currentFilter.value ?? "";
    const res = await usersApi.fetchUsers(data?.page || 1, data?.perPage || 10, filter);

    users.value = res.data as IAdminUser[];
    usersCount.value = parseInt(res.headers["x-total-count"] as string, 10);
  };

  const exportUsersToCsv = async (filter: string) => {
    const { data } = await usersApi.exportUsers(filter);
    return data;
  };

  const addUser = async (data: IAdminUserFormData) => {
    await usersApi.addUser(data);
  };

  const fetchUserById = async (id: string) => {
    const { data } = await usersApi.getUser(id);
    return data as IAdminUser;
  };

  const updateUser = async (data: IAdminUserFormData) => {
    const { id } = data;
    await usersApi.updateUser(id as string, data);
  };

  const deleteUser = async (id: string) => {
    await usersApi.deleteUser(id);
  };

  const resetUserPassword = async (id: string) => {
    const { data } = await usersApi.resetUserPassword(id);
    return data as string;
  };

  return {
    users,
    usersCount,
    currentFilter,
    setFilter,
    fetchUsersList,
    exportUsersToCsv,
    addUser,
    fetchUserById,
    updateUser,
    deleteUser,
    resetUserPassword,
  };
});

export default useUsersStore;
