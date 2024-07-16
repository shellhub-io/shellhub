import { IUser, IUserPutSecurity, IUserUpdatePassword, IUserSignUp } from "@/interfaces/IUSer";
import { usersApi } from "../../api/http";

export const signUp = async (data : IUserSignUp) => usersApi.registerUser({
  name: data.name,
  email: data.email,
  username: data.username,
  password: data.password,
  email_marketing: data.emailMarketing || false,
});

export const postResendEmail = async (username : string) => usersApi.resendEmail({ username });

export const postRecoverPassword = async (username : string) => usersApi.recoverPassword({ username });

export const postValidationAccount = async (data : IUser) => usersApi.getValidateAccount(data.email, data.token);

export const putSecurity = async (data : IUserPutSecurity) => usersApi.setSessionRecord(data.id, { session_record: data.status });

export const getSecurity = async () => usersApi.getSessionRecord();

export const postUpdatePassword = async (data : IUserUpdatePassword) => usersApi.updateRecoverPassword(data.id, {
  token: data.token,
  password: data.password,
});

export const patchUserData = async (data : IUser) => usersApi.updateUser({
  name: data.name,
  username: data.username,
  email: data.email,
  recovery_email: data.recovery_email,
});

export const patchUserPassword = async (data : IUser) => usersApi.updateUser({
  name: data.name,
  username: data.username,
  email: data.email,
  recovery_email: data.recovery_email,
  current_password: data.currentPassword,
  password: data.newPassword,
});
