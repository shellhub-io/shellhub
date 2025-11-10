export type UserStatus = "confirmed" | "invited" | "not-confirmed";

export type UserAuthMethods = Array<"saml" | "local">;

export interface IAdminUser {
  id: string;
  status: UserStatus;
  max_namespaces: number;
  created_at: string;
  last_login: string;
  name: string;
  username: string;
  email: string;
  recovery_email?: string | null;
  mfa?: { enabled?: boolean } | null;
  namespacesOwned: number;
  preferences?: {
    auth_methods: UserAuthMethods;
  };
  email_marketing?: boolean | null;
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
