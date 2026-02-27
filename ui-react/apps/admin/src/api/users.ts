import apiClient from "./client";

export interface SignUpPayload {
  name: string;
  email: string;
  username: string;
  password: string;
  email_marketing: boolean;
  sig?: string;
}

// Only token and tenant are consumed by the sign-up flow. The backend returns
// a full UserAuthResponse (13 fields) but the remaining fields are hydrated
// later via fetchUser() once the user is redirected to the app.
export interface SignUpResponse {
  token?: string;
  tenant?: string;
}

export async function signUp(payload: SignUpPayload): Promise<SignUpResponse> {
  const { data } = await apiClient.post<SignUpResponse>("/api/register", payload);
  return data;
}

export async function resendEmail(username: string): Promise<void> {
  await apiClient.post("/api/user/resend_email", { username });
}

export async function validateAccount(email: string, token: string, signal?: AbortSignal): Promise<void> {
  await apiClient.get("/api/user/validation_account", { params: { email, token }, signal });
}
