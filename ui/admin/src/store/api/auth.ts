import { adminApi } from "@/api/http";

type UserLogin = {
  username: string;
  password: string;
};

const login = async (user: UserLogin) => adminApi.loginAdmin(user);
const getToken = async (tenant: string) => adminApi.getUserTokenAdmin(tenant);

export { login, getToken };
