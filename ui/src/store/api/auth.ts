import {
  IUserLogin, IMfaEnable, IMfaDisable,
  IMfaAuth,
  IMfaRecover, IMfaReset,
} from "@/interfaces/IUserLogin";
import { usersApi, mfaApi } from "@/api/http";

export const login = async (user: IUserLogin) => usersApi.login(user);

export const validateMFA = async (validation: IMfaAuth) => mfaApi.authMFA({
  token: validation.token,
  code: validation.code,
});

export const recoverMfa = async (validation: IMfaRecover) => mfaApi.mfaRecover({
  identifier: validation.identifier,
  recovery_code: validation.recovery_code,
});

export const requestResetMfa = async (email: string) => mfaApi.requestResetMFA({
  identifier: email,
});

export const generateMfa = async () => mfaApi.generateMFA();

export const enableMFA = async (mfa: IMfaEnable) => mfaApi.enableMFA({
  code: mfa.code,
  secret: mfa.secret,
  recovery_codes: mfa.recovery_codes,
});

export const disableMfa = async (validation: Partial<IMfaDisable>) => mfaApi.disableMFA({
  code: validation.code,
  recovery_code: validation.recovery_code,
});

export const resetMfa = async (validation: IMfaReset) => mfaApi.resetMFA(validation.id, {
  main_email_code: validation.main_email_code,
  recovery_email_code: validation.recovery_email_code,
});

export const deleteUser = async () => usersApi.deleteUser();

export const getUserInfo = async () => usersApi.getUserInfo();
