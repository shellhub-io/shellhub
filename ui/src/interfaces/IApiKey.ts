import { BasicRole } from "@/interfaces/INamespace";

enum ApiKeyExpireList {
  NUMBER_30 = 30,
  NUMBER_60 = 60,
  NUMBER_90 = 90,
  NUMBER_365 = 365,
  NUMBER_MINUS_1 = -1,
}
export interface IApiKey {
  id: string,
  tenant_id: string,
  name: string,
  role: BasicRole,
  expires_in: ApiKeyExpireList,
}

export interface IApiKeyCreate {
  name: string,
  role: string,
  expires_in: ApiKeyExpireList,
}

export interface IApiKeyEdit {
  key: string,
  name?: string,
  role: string,
}

export interface IApiKeyRemove {
  key: string,
}
