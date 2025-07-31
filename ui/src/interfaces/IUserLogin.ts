export interface IUserLogin {
  username: string;
  password: string;
}

export interface IMfaGenerate {
  link: string,
  secret: string,
  recovery_codes: Array<string>
}

export interface IMfaEnable {
  code: string,
  secret: string,
  recovery_codes: Array<string>
}

export interface IMfaAuth {
  token: string,
  code: string
}

export interface IMfaDisable {
  code: string,
  recovery_code: string
}

export interface IMfaRecover {
  identifier: string,
  recovery_code: string
}

export interface IMfaReset {
  id: string,
  main_email_code: string,
  recovery_email_code: string,
}
