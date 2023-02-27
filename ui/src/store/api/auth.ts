import { IUserLogin } from "@/interfaces/IUserLogin";
import { usersApi } from "../../api/http";
import http from "../helpers/http";

export const login = async (user: IUserLogin) => usersApi.login(user);

export const info = async () => http().get("/auth/user");
