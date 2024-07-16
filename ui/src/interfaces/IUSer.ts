export interface IUser {
  email: string;
  id: string;
  name: string;
  role: string;
  tenant: string;
  token: string;
  user: string;
  username: string;
  recovery_email: string;
  currentPassword: string;
  newPassword: string;
}

export interface IUserSignUp {
  email: string;
  name: string;
  password: string;
  username: string;
  emailMarketing?: boolean;
}

export interface IUserUpdatePassword {
  id: string;
  token: string;
  password: string;
}

export interface IUserPutSecurity {
  id: string;
  status: boolean;
}
