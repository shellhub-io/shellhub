import {
  IUser,
  IUserPutSessionRecording,
  IUserUpdatePassword,
  IUserSignUp,
  IUserSetup,
  IUserPatch,
  IUserPatchPassword,
} from "@/interfaces/IUser";
import { usersApi, systemApi } from "@/api/http";

export const signUp = async (data: IUserSignUp) => usersApi.registerUser({
  name: data.name,
  email: data.email,
  username: data.username,
  password: data.password,
  email_marketing: data.emailMarketing || false,
  sig: data.sig,
});

export const resendEmail = async (username: string) => usersApi.resendEmail({ username });

export const recoverPassword = async (username: string) => usersApi.recoverPassword({ username });

export const validateAccount = async (data: Pick<IUser, "email" | "token">) => usersApi.getValidateAccount(data.email, data.token);

export const setSessionRecordStatus = async (data: IUserPutSessionRecording) => usersApi.setSessionRecord(
  data.id,
  { session_record: data.status },
);

export const getSessionRecordStatus = async () => usersApi.checkSessionRecord();

export const updatePassword = async (data: IUserUpdatePassword) => usersApi.updateRecoverPassword(data.id, {
  token: data.token,
  password: data.password,
});

export const patchUserData = async (data: IUserPatch) => usersApi.updateUser(data);

export const patchUserPassword = async (data: IUserPatchPassword) => usersApi.updateUser({
  ...data,
  current_password: data.currentPassword,
  password: data.newPassword,
});

export const getPremiumContent = async () => {
  const response = await fetch("https://static.shellhub.io/premium-features.v1.json");
  const data = await response.json();
  return data;
};

export const getSamlLink = async () => usersApi.getSamlAuthUrl();

export const setup = async (data: IUserSetup) => systemApi.setup(data.sign, {
  name: data.name,
  username: data.username,
  email: data.email,
  password: data.password,
});

export const getInfo = async () => systemApi.getInfo();
