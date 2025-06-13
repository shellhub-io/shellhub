export type UserStatus = "confirmed" | "invited" | "not-confirmed";

export type UserAuthMethods = Array<"saml" | "local">;

export interface IUser {
  id: string;
  namespaces: number;
  max_namespaces: number;
  status: UserStatus;
  created_at: string;
  last_login: string;
  name: string;
  email: string;
  username: string;
  password: string;
  preferences: {
    auth_methods: UserAuthMethods;
  }
}
