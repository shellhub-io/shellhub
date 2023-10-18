import { IUserLogin, MfaValidation, MfaCode } from "@/interfaces/IUserLogin";
import { usersApi, mfaApi } from "../../api/http";

export const login = async (user: IUserLogin) => usersApi.login(user);

export const disableMfa = async () => mfaApi.disableMFA();

export const enableMFA = async (mfa: MfaValidation) => mfaApi.enableMFA({
  token_mfa: mfa.token_mfa,
  secret: mfa.secret,
  codes: mfa.codes,
});

export const validateMFA = async (validation: MfaCode) => mfaApi.code(validation);

export const validateRecoveryCodes = async (validation: MfaCode) => mfaApi.recoveryCodes(validation);

export const generateMfa = async () => mfaApi.generateMFA();

export const info = async () => usersApi.getUserInfo();
