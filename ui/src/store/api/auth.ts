import { IUserLogin, MfaValidation, MfaCode,
  ApiKeyValidation, ApiKeyEdit,
  ApiKeyRemove } from "@/interfaces/IUserLogin";
import { usersApi, mfaApi, apiKeysApi } from "../../api/http";

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
