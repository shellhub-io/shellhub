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
