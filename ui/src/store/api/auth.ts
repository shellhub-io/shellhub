import { IUserLogin, MfaValidation, MfaDisable,
  MfaAuth,
  MfaRecover, MfaReset } from "@/interfaces/IUserLogin";
import { usersApi, mfaApi } from "@/api/http";

export const login = async (user: IUserLogin) => usersApi.login(user);

export const validateMFA = async (validation: MfaAuth) => mfaApi.authMFA({
  token: validation.token,
  code: validation.code,
});

export const validateRecoveryCodes = async (validation: MfaRecover) => mfaApi.mfaRecover({
  identifier: validation.identifier,
  recovery_code: validation.recovery_code,
});

export const reqResetMfa = async (id: string) => mfaApi.requestResetMFA({
  identifier: id,
});

export const generateMfa = async () => mfaApi.generateMFA();

export const enableMFA = async (mfa: MfaValidation) => mfaApi.enableMFA({
  code: mfa.code,
  secret: mfa.secret,
  recovery_codes: mfa.recovery_codes,
});

export const disableMfa = async (validation: MfaDisable) => mfaApi.disableMFA({
  code: validation.code,
  recovery_code: validation.recovery_code,
});

export const resetMfa = async (validation: MfaReset) => mfaApi.resetMFA(validation.id, {
  main_email_code: validation.main_email_code,
  recovery_email_code: validation.recovery_email_code,
});

export const deleteUser = async () => usersApi.deleteUser();

export const info = async () => usersApi.getUserInfo();
