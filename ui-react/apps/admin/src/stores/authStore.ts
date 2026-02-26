import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import type { AxiosResponse } from "axios";
import {
  getAuthUser,
  updateUser,
  updatePassword as apiUpdatePassword,
} from "../api/auth";
import {
  validateMfa,
  recoverMfa,
  requestMfaReset,
  completeMfaReset,
} from "../api/mfa";
import type { LoginResponse } from "../types/mfa";
import apiClient from "../api/client";

interface AuthState {
  token: string | null;
  user: string | null;
  userId: string | null;
  email: string | null;
  username: string | null;
  recoveryEmail: string | null;
  tenant: string | null;
  role: string | null;
  name: string | null;
  loading: boolean;
  error: string | null;
  mfaEnabled: boolean;
  mfaToken: string | null;
  mfaRecoveryExpiry: number | null;
  mfaResetUserId: string | null;
  mfaResetIdentifier: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  fetchUser: () => Promise<void>;
  setSession: (data: { token: string; tenant: string; role?: string }) => void;
  updateProfile: (data: {
    name?: string;
    username?: string;
    email?: string;
    recovery_email?: string;
  }) => Promise<void>;
  updatePassword: (
    currentPassword: string,
    newPassword: string,
  ) => Promise<void>;
  loginWithMfa: (code: string) => Promise<void>;
  recoverWithCode: (code: string, identifier?: string) => Promise<void>;
  requestMfaReset: (identifier: string) => Promise<void>;
  completeMfaReset: (mainEmailCode: string, recoveryEmailCode: string) => Promise<void>;
  updateMfaStatus: (enabled: boolean) => void;
  setMfaToken: (token: string) => void;
}

const initialState = {
  token: null,
  user: null,
  userId: null,
  email: null,
  username: null,
  recoveryEmail: null,
  tenant: null,
  role: null,
  name: null,
  loading: false,
  error: null,
  mfaEnabled: false,
  mfaToken: null,
  mfaRecoveryExpiry: null,
  mfaResetUserId: null,
  mfaResetIdentifier: null,
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      ...initialState,

      login: async (username: string, password: string) => {
        set({ loading: true, error: null });
        try {
          // Make direct API call to access response headers
          const response: AxiosResponse<LoginResponse> = await apiClient.post<LoginResponse>(
            "/api/login",
            { username, password }
          );

          // Check for MFA token in response headers (with null safety)
          const mfaToken = response.headers["x-mfa-token"] as string | undefined;

          if (mfaToken) {
            // MFA required - store token temporarily, don't persist
            set({
              mfaToken,
              mfaEnabled: true,
              user: username,
              loading: false,
            });
            return;
          }

          // Normal login without MFA
          const data = response.data;
          set({
            token: data.token,
            user: data.user,
            userId: data.id,
            email: data.email,
            tenant: data.tenant,
            name: data.name,
            mfaEnabled: data.mfa || false,
            loading: false,
          });
        } catch {
          // Check if MFA token was set by interceptor (401 with x-mfa-token)
          const currentState = get();
          if (currentState.mfaToken) {
            // MFA required - not an error, navigation will be handled by Login component
            // IMPORTANT: Set user field so recovery pages can access it
            set({
              loading: false,
              user: username,
              mfaEnabled: true,
            });
            return;
          }
          set({ loading: false, error: "Invalid username or password" });
        }
      },

      logout: () => {
        set(initialState);
        // Clear persisted session data from localStorage
        localStorage.removeItem("shellhub-session");
      },

      fetchUser: async () => {
        try {
          const user = await getAuthUser();
          set({
            user: user.user,
            username: user.user,
            userId: user.id,
            email: user.email,
            recoveryEmail: user.recovery_email,
            name: user.name,
            tenant: user.tenant,
            mfaEnabled: user.mfa || false,
          });
        } catch {
          /* session expired — interceptor handles redirect */
        }
      },

      setSession: ({ token, tenant, role }) => {
        set({ token, tenant, role: role ?? get().role });
      },

      updateProfile: async (data) => {
        await updateUser(data);
        await get().fetchUser();
      },

      updatePassword: async (currentPassword, newPassword) => {
        await apiUpdatePassword(currentPassword, newPassword);
      },

      loginWithMfa: async (code: string) => {
        const { mfaToken } = get();
        if (!mfaToken) {
          throw new Error("No MFA token available");
        }

        set({ loading: true, error: null });
        try {
          const data = await validateMfa({ token: mfaToken, code });
          set({
            token: data.token,
            user: data.user,
            userId: data.id,
            email: data.email,
            tenant: data.tenant,
            name: data.name,
            mfaToken: null, // Clear temporary token
            mfaEnabled: true,
            loading: false,
          });
        } catch {
          set({ loading: false, error: "Invalid verification code" });
          throw new Error("Invalid verification code");
        }
      },

      recoverWithCode: async (code: string, identifier?: string) => {
        // Try to get identifier from parameter first, then fall back to store
        const username = identifier || get().user || get().username;
        if (!username) {
          set({ error: "Username or email is required" });
          throw new Error("Username or email is required");
        }

        set({ loading: true, error: null });
        try {
          const { data, expiresAt } = await recoverMfa({
            identifier: username,
            recovery_code: code,
          });

          // Parse expiry time with validation
          let expiryValue: number | null = null;
          if (expiresAt) {
            const parsed = parseInt(expiresAt, 10);
            expiryValue = !isNaN(parsed) ? parsed : null;
          }

          set({
            token: data.token,
            user: data.user,
            userId: data.id,
            email: data.email,
            tenant: data.tenant,
            name: data.name,
            mfaEnabled: true,
            mfaRecoveryExpiry: expiryValue,
            loading: false,
          });
        } catch {
          set({ loading: false, error: "Invalid recovery code or username" });
          throw new Error("Invalid recovery code or username");
        }
      },

      requestMfaReset: async (identifier: string) => {
        set({ loading: true, error: null });
        try {
          const userId = await requestMfaReset(identifier);
          set({
            mfaResetUserId: userId,
            mfaResetIdentifier: identifier,
            loading: false,
          });
        } catch {
          set({ loading: false, error: "Unable to send reset emails. Please check your identifier." });
          throw new Error("Reset request failed");
        }
      },

      completeMfaReset: async (mainEmailCode: string, recoveryEmailCode: string) => {
        const { mfaResetUserId } = get();
        if (!mfaResetUserId) {
          set({ error: "Invalid reset session. Please start over." });
          throw new Error("No user ID available");
        }

        set({ loading: true, error: null });
        try {
          const data = await completeMfaReset(mfaResetUserId, {
            main_email_code: mainEmailCode,
            recovery_email_code: recoveryEmailCode,
          });

          // Successful reset = authenticated, same as login
          set({
            token: data.token,
            user: data.user,
            userId: data.id,
            email: data.email,
            tenant: data.tenant,
            name: data.name,
            mfaEnabled: data.mfa || false,
            mfaResetUserId: null,
            mfaResetIdentifier: null,
            loading: false,
          });
        } catch {
          set({ loading: false, error: "Invalid verification codes. Please check and try again." });
          throw new Error("Invalid codes");
        }
      },

      updateMfaStatus: (enabled: boolean) => {
        set({ mfaEnabled: enabled });
      },

      setMfaToken: (token: string) => {
        set({ mfaToken: token });
      },
    }),
    {
      name: "shellhub-session",
      version: 1,
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        userId: state.userId,
        email: state.email,
        username: state.username,
        recoveryEmail: state.recoveryEmail,
        tenant: state.tenant,
        role: state.role,
        name: state.name,
        mfaEnabled: state.mfaEnabled,
        // Do NOT persist: mfaToken, mfaRecoveryExpiry, mfaResetUserId, mfaResetIdentifier
      }),
    },
  ),
);
