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
