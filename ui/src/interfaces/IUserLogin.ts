export interface IUserLogin {
  username: string;
  password: string;
}

export interface MfaValidation {
  code: string,
  secret: string,
  recovery_codes: Array<string>
}

export interface MfaAuth {
  token: string,
  code: string
}

export interface MfaDisable {
  code: string,
  recovery_code: string
}

export interface MfaRecover {
  identifier: string,
  recovery_code: string
}

export interface MfaReset {
  id: string,
  main_email_code: string,
  recovery_email_code: string,
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
