import { IUserLogin, MfaValidation, MfaDisable,
  ApiKeyValidation, ApiKeyEdit,
  ApiKeyRemove,
  MfaAuth,
  MfaRecover, MfaReset } from "@/interfaces/IUserLogin";
import { usersApi, mfaApi, apiKeysApi } from "../../api/http";

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

export const info = async () => usersApi.getUserInfo();

export const generateApiKey = async (data: ApiKeyValidation) => apiKeysApi.createApiKey(data.tenant, {
  name: data.name,
  expires_at: data.expires_at,
});

export const getApiKey = async (
  tenant: string,
  page: number,
  perPage: number,
  sortStatusString : "asc" | "desc" | "",
  sortStatusField : string | undefined,
) => {
  if (sortStatusField && sortStatusString) {
    return apiKeysApi.listApiKey(
      tenant,
      page,
      perPage,
      sortStatusString,
      sortStatusField,
    );
  }

  return apiKeysApi.listApiKey(tenant, page, perPage);
};
export const removeApiKey = async (data: ApiKeyRemove) => apiKeysApi.deleteApiKey(data.tenant, data.id);

export const renameApiKey = async (data: ApiKeyEdit) => apiKeysApi.updateApiKey(data.tenant, data.id, { name: data.name });
