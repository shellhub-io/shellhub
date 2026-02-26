import apiClient from "./client";

interface LoginPayload {
  username: string;
  password: string;
}

interface LoginResponse {
  token: string;
  user: string;
  name: string;
  id: string;
  tenant: string;
  email: string;
}

interface UserResponse {
  id: string;
  name: string;
  user: string;
  email: string;
  recovery_email: string;
  tenant: string;
  mfa?: boolean;
}

export async function login(payload: LoginPayload): Promise<LoginResponse> {
  const { data } = await apiClient.post<LoginResponse>("/api/login", payload);
  return data;
}

export async function getAuthUser(): Promise<UserResponse> {
  const { data } = await apiClient.get<UserResponse>("/api/auth/user");
  return data;
}

export async function updateUser(data: {
  name?: string;
  username?: string;
  email?: string;
  recovery_email?: string;
}) {
  await apiClient.patch("/api/users", data);
}

export async function updatePassword(
  currentPassword: string,
  newPassword: string,
) {
  await apiClient.patch("/api/users", {
    current_password: currentPassword,
    password: newPassword,
  });
}
