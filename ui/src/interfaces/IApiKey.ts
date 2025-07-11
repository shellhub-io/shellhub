enum ApiKeyExpireList {
  NUMBER_30 = 30,
  NUMBER_60 = 60,
  NUMBER_90 = 90,
  NUMBER_365 = 365,
  NUMBER_MINUS_1 = -1,
}
export interface IApiKey {
  tenant_id: string,
  name: string,
  role: string,
  expires_at: ApiKeyExpireList,
}

export interface ApiKeyEdit {
  key: string,
  name: string,
  role: string,
}

export interface ApiKeyRemove {
  key: string,
}
