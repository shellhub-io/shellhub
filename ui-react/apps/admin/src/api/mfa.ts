import apiClient from "./client";
import type {
  MfaGenerateResponse,
  MfaEnableRequest,
  MfaAuthRequest,
  MfaDisableRequest,
  MfaRecoverRequest,
  MfaResetRequest,
  LoginResponse,
} from "../types/mfa";

// Generate QR code and recovery codes
export async function generateMfa(): Promise<MfaGenerateResponse> {
  const { data } = await apiClient.get<MfaGenerateResponse>(
    "/api/user/mfa/generate"
  );
  return data;
}

// Enable MFA with verification
export async function enableMfa(payload: MfaEnableRequest): Promise<void> {
  await apiClient.put("/api/user/mfa/enable", payload);
}

// Validate MFA code after password login
export async function validateMfa(
  payload: MfaAuthRequest
): Promise<LoginResponse> {
  const { data } = await apiClient.post<LoginResponse>(
    "/api/user/mfa/auth",
    payload
  );
  return data;
}

// Disable MFA
export async function disableMfa(payload: MfaDisableRequest): Promise<void> {
  await apiClient.put("/api/user/mfa/disable", payload);
}

// Recover account with recovery code
export async function recoverMfa(
  payload: MfaRecoverRequest
): Promise<{ data: LoginResponse; expiresAt: string }> {
  const response = await apiClient.post<LoginResponse>(
    "/api/user/mfa/recover",
    payload
  );
  return {
    data: response.data,
    expiresAt: (response.headers["x-expires-at"] as string | undefined) || "",
  };
}

// Request MFA reset via email
export async function requestMfaReset(identifier: string): Promise<string> {
  await apiClient.post("/api/user/mfa/reset", { identifier });
  // User ID will come from the email link, not from this response
  return identifier;
}

// Complete MFA reset with email codes
export async function completeMfaReset(
  userId: string,
  payload: MfaResetRequest
): Promise<LoginResponse> {
  const { data } = await apiClient.put<LoginResponse>(
    `/api/user/mfa/reset/${userId}`,
    payload
  );
  return data;
}
