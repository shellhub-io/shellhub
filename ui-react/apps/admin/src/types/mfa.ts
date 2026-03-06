export interface MfaGenerateResponse {
  link: string; // otpauth:// URL for QR code
  secret: string; // TOTP secret for manual entry
  recovery_codes: string[]; // 6 recovery codes
}

export interface MfaEnableRequest {
  code: string; // 6-digit TOTP code
  secret: string;
  recovery_codes: string[];
}

export interface MfaAuthRequest {
  token: string; // x-mfa-token from login response
  code: string; // 6-digit TOTP code
}

export interface MfaDisableRequest {
  code?: string; // TOTP code OR
  recovery_code?: string; // recovery code OR
  main_email_code?: string; // email verification code (requires recovery_email_code)
  recovery_email_code?: string; // email verification code (requires main_email_code)
}

export interface MfaRecoverRequest {
  identifier: string; // username or email
  recovery_code: string;
}

export interface MfaResetRequest {
  main_email_code: string;
  recovery_email_code: string;
}

export interface LoginResponse {
  token: string;
  user: string;
  name: string;
  id: string;
  tenant: string;
  email: string;
  mfa?: boolean; // Whether user has MFA enabled
}
