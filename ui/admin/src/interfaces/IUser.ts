export type UserStatus = "confirmed" | "invited" | "not-confirmed";

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
    auth_methods: Array<"saml" | "local">;
  }
}
