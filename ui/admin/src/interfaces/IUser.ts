export type UserStatus = "confirmed" | "invited" | "not-confirmed";

export type UserAuthMethods = Array<"saml" | "local">;

export interface IAdminUser {
  id: string;
  namespacesOwned: number;
  max_namespaces: number;
  status: UserStatus;
  created_at: string;
  last_login: string;
  name: string;
  username: string;
  email: string;
  recovery_email: string;
  mfa: {
    enabled: boolean;
  }
  preferences: {
    auth_methods: UserAuthMethods;
  }
}

export interface IAdminUserFormData {
  name: string;
  email: string;
  username: string;
  password: string;
  max_namespaces?: number;
  confirmed?: boolean;
  status: UserStatus;
  id?: string;
}
