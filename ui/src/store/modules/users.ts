import { defineStore } from "pinia";
import { ref } from "vue";
import * as usersApi from "../api/users";
import useAuthStore from "@/store/modules/auth";
import { IUser, IUserPatch, IUserPatchPassword, IUserSetup, IUserSignUp, IUserSystemInfo, IUserUpdatePassword } from "@/interfaces/IUser";

const useUsersStore = defineStore("users", () => {
  const showPaywall = ref(false);
  const signUpToken = ref<string>();
  const systemInfo = ref<IUserSystemInfo>({} as IUserSystemInfo);

  const signUp = async (data: IUserSignUp) => {
    // OpenAPI typing issue workaround
    const { data: user } = await usersApi.signUp(data) as unknown as { data: { token?: string } };

    if (!user.token) return false;

    useAuthStore().persistAuth(user);
    return user.token;
  };

  const patchData = async (data: IUserPatch) => {
    await usersApi.patchUserData(data);
  };

  const setup = async (data: IUserSetup) => {
    await usersApi.setup(data);
  };

  const patchPassword = async (data: IUserPatchPassword) => {
    await usersApi.patchUserPassword(data);
  };

  const resendEmail = async (username: string) => {
    await usersApi.resendEmail(username);
  };

  const recoverPassword = async (username: string) => {
    await usersApi.recoverPassword(username);
  };

  const validateAccount = async (data: Pick<IUser, "email" | "token">) => {
    await usersApi.validateAccount(data);
  };

  const updatePassword = async (data: IUserUpdatePassword) => {
    await usersApi.updatePassword(data);
  };

  const getPremiumContent = async () => usersApi.getPremiumContent();

  const fetchSystemInfo = async () => {
    const response = await usersApi.getInfo();
    systemInfo.value = response.data as IUserSystemInfo;
  };

  const getSamlUrl = async () => {
    const response = await usersApi.getSamlLink();
    return response.data.url;
  };

  return {
    showPaywall,
    signUpToken,
    systemInfo,

    signUp,
    patchData,
    setup,
    patchPassword,
    resendEmail,
    recoverPassword,
    validateAccount,
    updatePassword,
    getPremiumContent,
    fetchSystemInfo,
    getSamlUrl,
  };
});

export default useUsersStore;
