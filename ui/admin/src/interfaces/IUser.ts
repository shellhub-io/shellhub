export interface IUser {
  id: string;
  namespaces: number;
  status: "confirmed" | "invited" | "not-confirmed";
  created_at: string;
  last_login: string;
  name: string;
  email: string;
  username: string;
  password: string;
}
