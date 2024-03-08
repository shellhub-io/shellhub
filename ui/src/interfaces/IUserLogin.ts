export interface IUserLogin {
  username: string;
  password: string;
}

export interface MfaValidation {
  token_mfa: string,
  secret: string,
  codes: Array<string>
}

export interface MfaCode {
  code: string
}

enum ApiKeyExpireList {
  NUMBER_30 = 30,
  NUMBER_60 = 60,
  NUMBER_90 = 90,
  NUMBER_365 = 365,
  NUMBER_MINUS_1 = -1,
}
export interface ApiKeyValidation {
  tenant: string,
  name: string,
  expires_at: ApiKeyExpireList,
}

export interface ApiKeyEdit {
  tenant: string,
  name: string,
  id: string,
}

export interface ApiKey {
  name: string,
  expires_at: string,
}

export interface ApiKeyRemove {
  tenant: string,
  id: string,
}
export interface ApiKeyGetValidation {
  tenant: string,
  page : number,
  perPage: number,
  sortStatusField : string | undefined,
  sortStatusString : "asc" | "desc" | "",
}
