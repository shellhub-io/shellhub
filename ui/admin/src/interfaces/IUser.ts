export interface IUser {
  id: string;
  auth_methods: Array<string>;
  namespaces: number;
  confirmed: boolean;
  created_at: string;
  last_login: string;
  name: string;
  email: string;
  username: string;
  password: string;
}
